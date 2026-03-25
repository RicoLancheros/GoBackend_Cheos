package repository

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const notificationsCollection = "notifications"

type NotificationRepository struct {
	db *database.FirebaseClient
}

func NewNotificationRepository(db *database.FirebaseClient) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create persiste una nueva notificación en Firestore
func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	n.ID = uuid.New()
	n.CreatedAt = time.Now()
	n.Status = models.NotificationUnread

	log.Printf("[NotifRepo] Create notifID=%s userID=%s orden=%s estado=%s",
		n.ID, n.UserID, n.OrderNum, n.OrderStatus)

	_, err := r.db.Firestore.Collection(notificationsCollection).
		Doc(n.ID.String()).
		Set(ctx, n)
	if err != nil {
		log.Printf("[NotifRepo] ERROR Create: %v", err)
	} else {
		log.Printf("[NotifRepo] OK Create notifID=%s", n.ID)
	}
	return err
}

// GetByUserID retorna las últimas 50 notificaciones de un usuario, sin ordenamiento
// para evitar requerir índice compuesto — ordenamos en memoria
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Notification, error) {
	log.Printf("[NotifRepo] GetByUserID userID=%s", userID)

	// Sin OrderBy para no necesitar índice compuesto — ordenamos en Go
	iter := r.db.Firestore.Collection(notificationsCollection).
		Where("user_id", "==", userID).
		Limit(50).
		Documents(ctx)
	defer iter.Stop()

	var list []*models.Notification
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("[NotifRepo] ERROR iterando docs: %v", err)
			return nil, err
		}
		var n models.Notification
		if err := doc.DataTo(&n); err != nil {
			log.Printf("[NotifRepo] ERROR parseando doc %s: %v", doc.Ref.ID, err)
			continue
		}
		list = append(list, &n)
	}

	// Ordenar por created_at desc en memoria
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].CreatedAt.After(list[i].CreatedAt) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	log.Printf("[NotifRepo] OK GetByUserID — %d notificaciones", len(list))
	return list, nil
}

// MarkRead marca una notificación específica como leída verificando que pertenezca al usuario
func (r *NotificationRepository) MarkRead(ctx context.Context, notifID uuid.UUID, userID uuid.UUID) error {
	doc, err := r.db.Firestore.Collection(notificationsCollection).
		Doc(notifID.String()).Get(ctx)
	if err != nil {
		return err
	}

	var n models.Notification
	if err := doc.DataTo(&n); err != nil {
		return err
	}
	if n.UserID != userID {
		return nil
	}

	now := time.Now()
	_, err = r.db.Firestore.Collection(notificationsCollection).
		Doc(notifID.String()).
		Update(ctx, []firestore.Update{
			{Path: "status", Value: string(models.NotificationRead)},
			{Path: "read_at", Value: now},
		})
	return err
}

// MarkAllRead marca todas las notificaciones no leídas de un usuario como leídas (batch)
// Sin Where en "status" para evitar índice compuesto — filtramos en Go
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	iter := r.db.Firestore.Collection(notificationsCollection).
		Where("user_id", "==", userID).
		Documents(ctx)
	defer iter.Stop()

	now := time.Now()
	batch := r.db.Firestore.Batch()
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// Filtrar en Go — solo actualizar las no leídas
		var n models.Notification
		if err := doc.DataTo(&n); err != nil {
			continue
		}
		if n.Status != models.NotificationUnread {
			continue
		}

		batch.Update(doc.Ref, []firestore.Update{
			{Path: "status", Value: string(models.NotificationRead)},
			{Path: "read_at", Value: now},
		})
		count++

		if count == 499 {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = r.db.Firestore.Batch()
			count = 0
		}
	}

	if count > 0 {
		_, err := batch.Commit(ctx)
		return err
	}
	return nil
}

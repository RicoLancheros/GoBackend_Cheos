package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/cheoscafe/backend/internal/config"
	"google.golang.org/api/option"
)

type FirebaseClient struct {
	App       *firebase.App
	Auth      *auth.Client
	Firestore *firestore.Client
	Storage   *storage.Client
	ctx       context.Context
}

func NewFirebaseConnection(cfg *config.Config) (*FirebaseClient, error) {
	ctx := context.Background()

	// Configurar opciones de Firebase
	opt := option.WithCredentialsFile(cfg.FirebaseCredentialsPath)

	// Configuración de la app de Firebase
	firebaseConfig := &firebase.Config{
		ProjectID:     cfg.FirebaseProjectID,
		StorageBucket: cfg.FirebaseStorageBucket,
	}

	// Inicializar app de Firebase
	app, err := firebase.NewApp(ctx, firebaseConfig, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	// Inicializar cliente de Auth
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase auth: %v", err)
	}

	// Inicializar cliente de Firestore
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing firestore: %v", err)
	}

	// Inicializar cliente de Storage (opcional)
	var storageClient *storage.Client
	if cfg.FirebaseStorageBucket != "" {
		storageClient, err = storage.NewClient(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("error initializing storage: %v", err)
		}
	}

	return &FirebaseClient{
		App:       app,
		Auth:      authClient,
		Firestore: firestoreClient,
		Storage:   storageClient,
		ctx:       ctx,
	}, nil
}

func (fc *FirebaseClient) Close() error {
	if fc.Firestore != nil {
		return fc.Firestore.Close()
	}
	return nil
}

func (fc *FirebaseClient) GetContext() context.Context {
	return fc.ctx
}

// Collection devuelve una referencia a una colección de Firestore
func (fc *FirebaseClient) Collection(name string) *firestore.CollectionRef {
	return fc.Firestore.Collection(name)
}

// RunTransaction ejecuta una transacción de Firestore
func (fc *FirebaseClient) RunTransaction(ctx context.Context, f func(context.Context, *firestore.Transaction) error) error {
	return fc.Firestore.RunTransaction(ctx, f)
}

// Batch crea un nuevo batch para operaciones múltiples
func (fc *FirebaseClient) Batch() *firestore.WriteBatch {
	return fc.Firestore.Batch()
}
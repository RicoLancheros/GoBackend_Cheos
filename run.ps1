# Script para ejecutar el backend de Cheos Café
Write-Host "Iniciando Cheos Café Backend..." -ForegroundColor Green
Write-Host ""

# Cambiar al directorio del proyecto
Set-Location $PSScriptRoot

# Ejecutar el servidor
go run cmd/api/main.go

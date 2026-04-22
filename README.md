# 🐳 Docker Visual Backend

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-CC%20BY--NC%204.0-orange?style=for-the-badge)

**Docker Visual Backend** es el motor de procesamiento y gestión para la plataforma Docker Visual. Escrito en Go, este servicio interactúa directamente con el Docker Engine para proporcionar una API robusta y segura que permite la orquestación y monitoreo de contenedores, redes y volúmenes.

---

## 🚀 Instalación y Configuración

Sigue estos pasos para configurar el entorno de desarrollo:

### Requisitos Previos
- **Go**: 1.21 o superior.
- **Docker**: Debe estar instalado y el daemon en ejecución.
- **Git**: Para el control de versiones.

### Pasos de Instalación

1.  **Clonar el repositorio:**
    ```bash
    git clone <url-del-repositorio>
    cd docker-visual-backend
    ```

2.  **Configurar variables de entorno:**
    Copia el archivo de ejemplo y ajusta los valores según tu entorno:
    ```bash
    cp .env.example .env
    ```
    Variables principales:
    - `PORT`: Puerto en el que correrá el servidor (default: 8080).
    - `CORS_ORIGINS`: Orígenes permitidos para el frontend.
    - `WORKSPACE_PATH`: Directorio donde se clonarán proyectos externos.

3.  **Instalar dependencias:**
    ```bash
    go mod download
    ```

4.  **Ejecutar el servidor:**
    ```bash
    go run cmd/server/main.go
    ```

---

## 🛠️ Uso y Endpoints

El backend expone una API RESTful bajo el prefijo `/api`. Algunos de los endpoints principales incluyen:

-   **Salud:** `GET /api/health` - Verifica el estado del servidor.
-   **Contenedores:** 
    - `GET /api/containers` - Listar todos los contenedores.
    - `POST /api/containers` - Crear un nuevo contenedor.
    - `POST /api/containers/:id/start` - Iniciar un contenedor.
-   **Proyectos:**
    - `POST /api/projects` - Desplegar proyectos desde repositorios Git.
-   **Visualización:**
    - `GET /api/graph` - Obtiene datos estructurados para la visualización en D3.js.

### Seguridad
Los endpoints protegidos requieren un token JWT que se obtiene tras el login exitoso en `/api/auth/login`.

---

## 💡 Motivación y Problema

Administrar múltiples contenedores Docker, redes y volúmenes a través de la terminal (CLI) puede volverse complejo y propenso a errores, especialmente al tratar de visualizar las interconexiones entre servicios. 

**Docker Visual Backend** resuelve esto al:
1.  Abstraer la complejidad de la API de Docker.
2.  Proporcionar una estructura de datos optimizada para visualizaciones gráficas.
3.  Permitir despliegues rápidos de proyectos basados en repositorios con gestión automática de túneles y variables de entorno.

---

## 👥 Créditos y Autores

Este proyecto ha sido desarrollado por:
- **Frey-r** - *Desarrollo Inicial y Arquitectura*

---

## 📄 Licencia

Este proyecto está bajo la licencia **Creative Commons Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)**.

**Puedes:**
- Compartir, copiar y redistribuir el material en cualquier medio o formato.
- Adaptar, remezclar, transformar y construir sobre el material.

**Bajo los siguientes términos:**
- **Atribución:** Debes dar crédito de manera adecuada.
- **No Comercial:** No puedes utilizar el material con fines comerciales o de lucro.

Para más detalles, consulta el archivo [LICENSE](LICENSE) o visita [creativecommons.org](https://creativecommons.org/licenses/by-nc/4.0/).

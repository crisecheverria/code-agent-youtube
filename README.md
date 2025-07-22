# Agente de Código AI - Tutorial

Un tutorial básico para crear un agente de código con inteligencia artificial que puede ejecutar comandos y manipular archivos.

## Descripción

Este proyecto es un agente de código AI que proporciona una interfaz web y de línea de comandos para interactuar con un asistente de IA que puede ejecutar herramientas del sistema. El agente está construido con una arquitectura modular que incluye:

- **Core**: Servidor web HTTP construido con Hono que maneja sesiones de conversación
- **TUI**: Interfaz de usuario de terminal escrita en Go para interacción por línea de comandos

## Características

- 🤖 **Asistente de IA**: Integración con Groq para capacidades de lenguaje natural
- 🔧 **Ejecución de Herramientas**: Ejecuta comandos bash, lee/escribe archivos, y maneja directorios
- 💬 **API de Conversación**: API RESTful completa para gestionar conversaciones
- 📡 **Streaming**: Soporte para respuestas en tiempo real mediante Server-Sent Events
- 🖥️ **TUI**: Interfaz de terminal interactiva para uso desde línea de comandos
- 📊 **Seguimiento de Tokens**: Monitoreo del uso de tokens para control de costos

## Estructura del Proyecto

```
code-agent/
├── packages/
│   ├── core/           # Servidor web principal (TypeScript/Bun)
│   │   ├── src/
│   │   │   ├── index.ts      # Servidor HTTP con endpoints
│   │   │   ├── session.ts    # Gestión de sesiones de conversación
│   │   │   ├── messages.ts   # Tipos y funciones de mensajes
│   │   │   ├── tools.ts      # Herramientas del sistema (bash, archivos)
│   │   │   └── groq.ts       # Cliente para API de Groq
│   │   └── package.json
│   └── tui/            # Interfaz de terminal (Go)
│       ├── main.go     # Cliente TUI
│       └── go.mod
├── bin/                # Binarios compilados
├── package.json        # Configuración del workspace
└── README.md
```

## Instalación

### Requisitos Previos

- [Bun](https://bun.sh/) para el runtime de JavaScript/TypeScript
- [Go](https://golang.org/) para la interfaz TUI
- Token de API de [Groq](https://groq.com/)

### Configuración

1. Clona el repositorio:
```bash
git clone <url-del-repositorio>
cd code-agent
```

2. Instala las dependencias:
```bash
bun install
```

3. Compila el proyecto:
```bash
bun run build
```

## Uso

### Servidor Web

Inicia el servidor de desarrollo:

```bash
bun run dev
```

El servidor estará disponible en `http://localhost:3000`

### Endpoints de la API

- `GET /health` - Verificación de estado del servidor
- `POST /session` - Inicializar nueva sesión de conversación
- `POST /message` - Enviar mensaje al agente
- `GET /stream` - Obtener respuestas en streaming
- `POST /tool` - Ejecutar herramienta específica
- `GET /conversation` - Obtener historial de conversación
- `GET /tools` - Listar herramientas disponibles
- `GET /tokens` - Obtener uso de tokens
- `DELETE /session` - Limpiar sesión actual

### Interfaz TUI

Ejecuta la interfaz de terminal:

```bash
./bin/tui
```

## Herramientas Disponibles

El agente incluye las siguientes herramientas del sistema:

- **bash**: Ejecutar comandos de shell
- **readFile**: Leer contenido de archivos
- **writeFile**: Escribir contenido a archivos
- **listFiles**: Listar archivos en directorios
- **makeDir**: Crear nuevos directorios

## Configuración

Para usar el agente, necesitas configurar una sesión con tus credenciales de Groq:

```json
{
  "groq": {
    "token": "tu-token-de-groq",
    "model": "llama-3.3-70b-versatile",
    "baseURL": "https://api.groq.com/openai"
  }
}
```

## Tecnologías Utilizadas

- **Backend**: TypeScript, Bun, Hono
- **TUI**: Go
- **IA**: Groq API con modelo Llama 3.3
- **Validación**: Zod
- **Arquitectura**: Monorepo con workspaces

## Desarrollo

### Scripts Disponibles

- `bun run dev` - Ejecutar servidor en modo desarrollo
- `bun run build` - Compilar todo el proyecto
- `bun run build:core` - Compilar solo el core
- `bun run build:tui` - Compilar solo la TUI

## Licencia

Este proyecto es un tutorial educativo para aprender sobre agentes de código con IA.
# Tiltfile for GOTHA Boilerplate
# Provides live reload development with Docker Compose

# Use Docker Compose for services
docker_compose('./docker/docker-compose.yml')

# Watch for changes in Go, Templ, and CSS files
# and trigger rebuilds accordingly

# Watch Go and Templ source files
dc_resource('app',
    trigger_mode=TRIGGER_MODE_AUTO,
)

# Run Tailwind CSS in watch mode locally
local_resource('tailwind-watch',
    serve_cmd='npm run css:watch',
    deps=['web/templates'],
)

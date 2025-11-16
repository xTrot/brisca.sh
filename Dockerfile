FROM alpine:latest
RUN apk add --no-cache bash
RUN apk add --no-cache openssh
WORKDIR /app
ADD brisca.sh /app/
ADD healthcheck /app/
ADD healthyloop /app/
EXPOSE 22

# Configurable env variables
# =============================================================================
# This is the game server's URL.
# ENV BRISCA_SERVER=http://brisca-server:8000
# This turns on debugging.
# ENV BRISCA_DEBUG=true
# This is the address the wish ssh server listens on.
# ENV BRISCA_HOST=0.0.0.0
# This is the port the wish ssh server listens on.
# ENV BRISCA_PORT=22
# =============================================================================

# Required volume
# =============================================================================
# /app/.ssh/
# This stores the secret ssh key for the server's ssh fingerprint.
# If not provided a new key will be generated and users will be prompted to 
#     distrust this change.
# =============================================================================

CMD ["/app/brisca.sh"]
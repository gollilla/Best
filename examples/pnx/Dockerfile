FROM eclipse-temurin:21-jre

WORKDIR /nukkit

# Download PowerNukkitX
RUN apt-get update && apt-get install -y curl && \
    curl -L -o pnx.jar https://github.com/PowerNukkitX/PowerNukkitX/releases/download/snapshot/powernukkitx.jar && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Create data directory
RUN mkdir -p /nukkit/data

# Copy config files (PNX 2.0 uses pnx.yml instead of nukkit.yml)
COPY pnx.yml /nukkit/pnx.yml
COPY ops.txt /nukkit/ops.txt

EXPOSE 19132/udp

# Start server
CMD ["java", "-Xms512M", "-Xmx1G", "-jar", "pnx.jar", "--language", "eng"]

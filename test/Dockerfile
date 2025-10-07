FROM debian:11

# manual testing:
#   build: docker build -f Dockerfile -t squid-test --network=host --no-cache .
#   run: docker run -it --rm --network=host --volume /tmp/squid/:/var/log/squid/ squid-tes

ENV SQUID_DIR_CACHE=/var/spool/squid \
    SQUID_DIR_LOG=/var/log/squid \
    SQUID_DIR_LIB=/var/lib/squid \
    SQUID_USER=proxy \
    SQUID_DH_SIZE=1024 \
    SQUID_SSLDB_SIZE=50MB \
    SQUID_DOCKER_LOGS=yes \
    SQUID_DOCKER_LOGS_CACHE=yes \
    SQUID_CERT_CN="/CN=Forward Proxy"
# "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=/CN=Forward Proxy"

RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y squid-openssl \
  && rm -rf /var/lib/apt/lists/* \
  && apt-get clean

COPY --chmod=640 --chown="$SQUID_USER":"$SQUID_USER" files/squid.conf /etc/squid/squid.conf
COPY --chmod=755 files/entrypoint.sh /sbin/entrypoint.sh

EXPOSE 3128/tcp
ENTRYPOINT ["/sbin/entrypoint.sh"]

#!/bin/bash

set -euo pipefail

SQUID_DIR_CONF="/etc/squid"
SQUID_DIR_SSL="${SQUID_DIR_LIB}/ssl"
SQUID_DIR_SSLDB="${SQUID_DIR_LIB}/ssl_db"

create_dir() {
  directory="$1"
  if ! [ -d "$directory" ]
  then
    mkdir -p "$directory"
  fi
  chown -R "$SQUID_USER":"$SQUID_USER" "$directory"
  chmod 750 "$directory"
}

create_missing_certs() {
  if ! [ -f "${SQUID_DIR_SSL}/squid.crt" ]
  then
    echo ''
    echo '### CREATING CERTIFICATES ###'
    openssl req -x509 -newkey rsa:4096 -keyout "${SQUID_DIR_SSL}/bump.key" -out "${SQUID_DIR_SSL}/bump.crt" -sha256 -days 3650 -nodes -subj "$SQUID_CERT_CN"
    openssl dhparam -outform PEM -out "${SQUID_DIR_SSL}/bump.dh.pem" "$SQUID_DH_SIZE"
  fi
}

create_missing_ssldb() {
  if ! [ -f "${SQUID_DIR_SSLDB}/index.txt" ]
  then
    echo ''
    echo '### CREATING SSL-DB ###'
    rm -rf "$SQUID_DIR_SSLDB"
    /usr/lib/squid/security_file_certgen -c -s "$SQUID_DIR_SSLDB" -M "$SQUID_SSLDB_SIZE"
    chown -R "$SQUID_USER":"$SQUID_USER" "${SQUID_DIR_SSLDB}"
    chmod 750 "$SQUID_DIR_SSLDB"
  fi
}

create_missing_logfile() {
  logfile="${SQUID_DIR_LOG}/$1"
  if ! [ -f "$logfile" ]
  then
    touch "$logfile"
    chown "$SQUID_USER":"$SQUID_USER" "$logfile"
    chmod 640 "$logfile"
  fi
}

create_dir "$SQUID_DIR_LOG"
create_dir "$SQUID_DIR_CACHE"
create_dir "$SQUID_DIR_SSL"
create_dir "$SQUID_DIR_LIB"
create_missing_certs
create_missing_ssldb

# docker logs
if [[ "$SQUID_DOCKER_LOGS" == "yes" ]]
then
  create_missing_logfile "access.log"
  create_missing_logfile "error.log"
  tail -F "${SQUID_DIR_LOG}/access.log" 2>/dev/null &
  tail -F "${SQUID_DIR_LOG}/error.log" 2>/dev/null &
  if [[ "$SQUID_DOCKER_LOGS_CACHE" == "yes" ]]
  then
    create_missing_logfile "cache.log"
    create_missing_logfile "store.log"
    tail -F "${SQUID_DIR_LOG}/cache.log" 2>/dev/null &
    tail -F "${SQUID_DIR_LOG}/store.log" 2>/dev/null &
  fi
fi

SQUID=$(which squid)
CONFIG="${SQUID_DIR_CONF}/squid.conf"

echo ''
echo '### DEBUG INFO ###'
echo "LIB DIR: '${SQUID_DIR_LIB}'"
echo "SSL DIR: '${SQUID_DIR_SSL}'"
echo "SSL-DB DIR: '${SQUID_DIR_SSLDB}'"
echo "LOG DIR: '${SQUID_DIR_LOG}'"
echo "CACHE DIR: '${SQUID_DIR_CACHE}'"
apt policy squid-openssl
$SQUID --version
#echo ''
#$SQUID --help || true

echo ''
echo '### CREATING MISSING CACHE DIRECTORIES ###'
$SQUID -N -f "$CONFIG" -z

echo ''
echo '### CHECKING CONFIG ###'
$SQUID -f "$CONFIG" -k parse

echo ''
echo '### STARTING SQUID ###'
exec $SQUID -f "$CONFIG" -NYCd 1 #$SQUID_EXTRA_ARGS

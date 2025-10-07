#!/usr/bin/env bash

set -euo pipefail

FAIL_FILE='/tmp/.failed'
RES_SUCCESS=200
RES_REDIRECT=301
RES_DENIED=403
RES_DENIED_TLS=000
TEST_DOMAIN='one.one.one.one'

if [[ "$UID" != "0" ]]
then
  echo 'ERROR: Script needs to be ran as root'
  exit 1
fi

echo "1.1.1.1 ${TEST_DOMAIN}" | tee -a /etc/hosts
echo "2606:4700:4700::1111 ${TEST_DOMAIN}" | tee -a /etc/hosts

function run_test() {
  url="$1"
  flags="$2"
  result="$3"

  res="$(curl -w '%{response_code}' -o /dev/null $flags "$url" 2>/dev/null)"
  if [[ "$res" != "$result" ]]
  then
    echo "ERROR: Response not as expected! Flags '${flags}' ${res} != ${result}"
    touch "$FAIL_FILE"
  fi

  # todo: make sure the log-entry in /tmp/fwd*.log also matches (it was forwarded)

  out="$(tail -n 2 '/tmp/squid/access.log')"
  if [[ "$result" == "$RES_SUCCESS" ]] && echo "$out" | grep -v "TCP_TUNNEL/${RES_SUCCESS}"
  then
    echo "ERROR: Response should be SUCCESSFUL '${out}'"
    touch "$FAIL_FILE"
  fi
  if [[ "$result" == "$RES_REDIRECT" ]] && echo "$out" | grep -v "TCP_MISS/${RES_REDIRECT}"
  then
    echo "ERROR: Response should be REDIRECT '${out}'"
    touch "$FAIL_FILE"
  fi
  if [[ "$result" == "$RES_DENIED" ]] || [[ "$result" == "$RES_DENIED_TLS" ]]
  then
    if echo "$out" | grep -v "TCP_DENIED/${RES_DENIED}"
    then
      echo "ERROR: Response should be DENIED '${out}'"
      touch "$FAIL_FILE"
    fi
  fi
}

run_test "http://${TEST_DOMAIN}" "--ipv4 --http1.0" "$RES_REDIRECT"
run_test "http://${TEST_DOMAIN}" "--ipv4 --http1.1" "$RES_REDIRECT"
run_test "http://${TEST_DOMAIN}" "--ipv4 --http2" "$RES_REDIRECT"
run_test "http://${TEST_DOMAIN}" "--ipv6 --http1.0" "$RES_REDIRECT"
run_test "http://${TEST_DOMAIN}" "--ipv6 --http1.1" "$RES_REDIRECT"
run_test "http://${TEST_DOMAIN}" "--ipv6 --http2" "$RES_REDIRECT"

run_test 'http://oxl.at' "--ipv4 --http1.1" "$RES_DENIED"

run_test "https://${TEST_DOMAIN}" "--ipv4 --http1.0" "$RES_SUCCESS"
run_test "https://${TEST_DOMAIN}" "--ipv4 --http1.1" "$RES_SUCCESS"
run_test "https://${TEST_DOMAIN}" "--ipv4 --http2" "$RES_SUCCESS"
run_test "https://${TEST_DOMAIN}" "--ipv6 --http1.0" "$RES_SUCCESS"
run_test "https://${TEST_DOMAIN}" "--ipv6 --http1.1" "$RES_SUCCESS"
run_test "https://${TEST_DOMAIN}" "--ipv6 --http2" "$RES_SUCCESS"

run_test 'https://oxl.at' "--ipv4 --http1.1" "$RES_DENIED_TLS"

if [ -f "$FAIL_FILE" ]
then
  exit 1
fi

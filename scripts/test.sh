#!/usr/bin/env bash

set -uo pipefail

FAIL_FILE='/tmp/.failed'
RES_SUCCESS=200
RES_REDIRECT=301
RES_DENIED=403
RES_DENIED_TLS=000
TEST_DOMAIN='one.one.one.one'
TEST_DOMAIN2='oxl.at'

echo 'INFO: Setting /etc/hosts'
echo "1.1.1.1 ${TEST_DOMAIN} ${TEST_DOMAIN2}" | sudo tee -a /etc/hosts
echo "2606:4700:4700::1111 ${TEST_DOMAIN} ${TEST_DOMAIN2}" | sudo tee -a /etc/hosts

function run_test() {
  nr="$1"
  url="$2"
  flags="$3"
  result="$4"
  success=1

  echo "${nr} | INFO: REQUEST TO ${url}"
  res="$(curl -w '%{response_code}' -o /dev/null $flags "$url" 2>/dev/null)"
  if [[ "$res" != "$result" ]]
  then
    echo "${nr} | ERROR: Response not as expected! ${res} != ${result}"
    touch "$FAIL_FILE"
    success=0
  fi

  # todo: make sure the log-entry in /tmp/fwd*.log also matches (it was forwarded)

  out="$(sudo tail -n 1 '/tmp/squid/access.log')"
  echo "$out"

  if [[ "$result" == "$RES_SUCCESS" ]] && echo "$out" | grep -qv "/${RES_SUCCESS} "
  then
    echo "${nr} | ERROR: Response should be SUCCESSFUL"
    touch "$FAIL_FILE"
    success=0
  fi
  if [[ "$result" == "$RES_REDIRECT" ]] && echo "$out" | grep -qv "/${RES_REDIRECT} "
  then
    echo "${nr} | ERROR: Response should be REDIRECT"
    touch "$FAIL_FILE"
    success=0
  fi
  if [[ "$result" == "$RES_DENIED" ]] || [[ "$result" == "$RES_DENIED_TLS" ]]
  then
    if echo "$out" | grep -qv "TCP_DENIED/${RES_DENIED} "
    then
      echo "${nr} | ERROR: Response should be DENIED"
      touch "$FAIL_FILE"
      success=0
    fi
  fi
  if [[ "$success" == "1" ]]
  then
    echo "${nr} | SUCCESS"
  fi
  echo ''
  sleep 1
}

run_test 1 "http://${TEST_DOMAIN}" "--ipv4 --http1.0" "$RES_REDIRECT"
run_test 2 "http://${TEST_DOMAIN}" "--ipv4 --http1.1" "$RES_REDIRECT"
run_test 3 "http://${TEST_DOMAIN}" "--ipv4 --http2" "$RES_REDIRECT"
run_test 4 "http://${TEST_DOMAIN}" "--ipv6 --http1.0" "$RES_REDIRECT"
run_test 5 "http://${TEST_DOMAIN}" "--ipv6 --http1.1" "$RES_REDIRECT"
run_test 6 "http://${TEST_DOMAIN}" "--ipv6 --http2" "$RES_REDIRECT"

run_test 7 "http://${TEST_DOMAIN2}" "--ipv4 --http1.1" "$RES_DENIED"

run_test 8 "https://${TEST_DOMAIN}" "--ipv4 --http1.0" "$RES_SUCCESS"
run_test 9 "https://${TEST_DOMAIN}" "--ipv4 --http1.1" "$RES_SUCCESS"
run_test 10 "https://${TEST_DOMAIN}" "--ipv4 --http2" "$RES_SUCCESS"
run_test 11 "https://${TEST_DOMAIN}" "--ipv6 --http1.0" "$RES_SUCCESS"
run_test 12 "https://${TEST_DOMAIN}" "--ipv6 --http1.1" "$RES_SUCCESS"
run_test 13 "https://${TEST_DOMAIN}" "--ipv6 --http2" "$RES_SUCCESS"

run_test 14 "https://${TEST_DOMAIN2}" "--ipv4 --http1.1" "$RES_DENIED_TLS"

if [ -f "$FAIL_FILE" ]
then
  echo '##########'
  echo '  FAILED  '
  echo '##########'
  exit 1
fi

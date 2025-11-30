#!/bin/sh

# This is a very simple approximation of an onboarding flow for
# a Kubernetes workload with kubespiffe. It uses the PSAT it has
# been given by the Kubernetes API with the /v1/svid endpoint of
# kubespiffd, and then tries to deserialise the response. If it
# successfully attests with the PSAT and there's a WorkloadRegistration
# CustomResource registered, then it will get an X509-SVID

export IS_SERVER
echo $IS_SERVER

echo "Workload booting..."
while true; do
  # Read PSAT token from projected volume
  TOKEN=$(cat /var/run/secrets/tokens/psat)
  echo "Using PSAT:"
  echo "$TOKEN"
              
  RESULT=$(curl -s -k -H "Authorization: Bearer $TOKEN" kubespiffed.kubespiffe.svc.cluster.local:8080/v1/svid)
  X509_SVID=$(echo "$RESULT" | jq -r '.x509_svid' | base64 -d)
  X509_SVID_KEY=$(echo "$RESULT" | jq -r '.x509_svid_key' | base64 -d)
  TRUST_BUNDLE=$(echo "$RESULT" | jq -r '.bundle' | base64 -d)

  if [ -n "$X509_SVID" ]; then
    echo "Obtained X509-SVID:"
    echo "$X509_SVID"

    echo "$X509_SVID" > /tmp/cert.pem
    echo "$X509_SVID_KEY"  > /tmp/key.pem
    echo "$TRUST_BUNDLE" > /tmp/cacert.pem

    echo "$X509_SVID" | openssl x509 -noout -ext subjectAltName
    echo ""
    break
  fi

  sleep 10
done

if [ "$IS_SERVER" = "true" ]; then
  echo "Starting server..."
  while true; do
    echo "Waiting for client connection..."
    openssl s_server \
      -accept 8080 \
      -cert /tmp/cert.pem \
      -key /tmp/key.pem \
      -CAfile /tmp/cacert.pem \
      -quiet \
      -www
    echo "Client disconnected, restarting..."
  done
fi
    
while true; do
  echo "Starting client..."
  openssl s_client \
    -connect server.default.svc.cluster.local:8080 \
    -cert /tmp/cert.pem \
    -key /tmp/key.pem \
    -CAfile /tmp/cacert.pem \
    -servername "server.default.svc.cluster.local" \
    -verify 1 2>/dev/null | openssl x509 -noout -subject || echo "mTLS failed"
  sleep 3
done


#!/bin/sh

# This is a very simple approximation of an onboarding flow for
# a Kubernetes workload with kubespiffe. It uses the PSAT it has
# been given by the Kubernetes API with the /v1/svid endpoint of
# kubespiffd, and then tries to deserialise the response. If it
# successfully attests with the PSAT and there's a WorkloadRegistration
# CustomResource registered, then it will get an X509-SVID

echo "Workload booting..."
while true; do
  # Read PSAT token from projected volume
  TOKEN=$(cat /var/run/secrets/tokens/psat)
  echo "Using PSAT:"
  echo "$TOKEN"
              
  RESULT=$(curl -s -k -H "Authorization: Bearer $TOKEN" kubespiffed.kubespiffe.svc.cluster.local:8080/v1/svid)
  echo "Obtained X509-SVID:"
  echo "$RESULT"

  echo "$RESULT" | openssl x509 -noout -ext subjectAltName
  echo ""
  sleep 20
done


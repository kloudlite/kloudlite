#!/bin/bash

# Test script for webhook validation and mutation

API_URL="http://localhost:8080"

echo "Testing Webhook Mutation - should add email label automatically"
echo "================================================================"

# Create a test AdmissionReview request for mutation
cat > /tmp/mutation-request.json <<EOF
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "request": {
    "uid": "test-$(date +%s)",
    "kind": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "kind": "User"
    },
    "resource": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "resource": "users"
    },
    "operation": "CREATE",
    "object": {
      "apiVersion": "platform.kloudlite.io/v1alpha1",
      "kind": "User",
      "metadata": {
        "name": "test-user",
        "namespace": "default"
      },
      "spec": {
        "email": "test@example.com",
        "displayName": "Test User",
        "roles": ["developer"]
      }
    }
  }
}
EOF

echo "Sending mutation webhook request..."
curl -X POST ${API_URL}/webhooks/mutate/users \
  -H "Content-Type: application/json" \
  -d @/tmp/mutation-request.json | jq '.'

echo ""
echo "Testing Webhook Validation - should validate email format"
echo "========================================================="

# Create a test AdmissionReview request for validation with invalid email
cat > /tmp/validation-request.json <<EOF
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "request": {
    "uid": "test-$(date +%s)",
    "kind": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "kind": "User"
    },
    "resource": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "resource": "users"
    },
    "operation": "CREATE",
    "object": {
      "apiVersion": "platform.kloudlite.io/v1alpha1",
      "kind": "User",
      "metadata": {
        "name": "test-user-invalid",
        "namespace": "default"
      },
      "spec": {
        "email": "invalid-email",
        "displayName": "Test User",
        "roles": []
      }
    }
  }
}
EOF

echo "Sending validation webhook request (should fail - invalid email and no roles)..."
curl -X POST ${API_URL}/webhooks/validate/users \
  -H "Content-Type: application/json" \
  -d @/tmp/validation-request.json | jq '.'

echo ""
echo "Testing Webhook Validation - valid request"
echo "==========================================="

# Create a valid request
cat > /tmp/validation-valid.json <<EOF
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "request": {
    "uid": "test-$(date +%s)",
    "kind": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "kind": "User"
    },
    "resource": {
      "group": "platform.kloudlite.io",
      "version": "v1alpha1",
      "resource": "users"
    },
    "operation": "CREATE",
    "object": {
      "apiVersion": "platform.kloudlite.io/v1alpha1",
      "kind": "User",
      "metadata": {
        "name": "valid-user",
        "namespace": "default"
      },
      "spec": {
        "email": "valid@example.com",
        "displayName": "Valid User",
        "roles": ["developer", "viewer"]
      }
    }
  }
}
EOF

echo "Sending validation webhook request (should pass)..."
curl -X POST ${API_URL}/webhooks/validate/users \
  -H "Content-Type: application/json" \
  -d @/tmp/validation-valid.json | jq '.'

# Clean up
rm -f /tmp/mutation-request.json /tmp/validation-request.json /tmp/validation-valid.json

echo ""
echo "Test complete!"
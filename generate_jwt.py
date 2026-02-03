import jwt
import time

APP_ID = "2782987"
PRIVATE_KEY_PATH = "/home/matthew/Downloads/workflowduplicationcost.2026-02-02.private-key.pem"

with open(PRIVATE_KEY_PATH, 'rb') as f:
    private_key_bytes = f.read()

# Define the claims for the JWT
now = int(time.time())
payload = {
    'iat': now,       # Issued at time
    'exp': now + 300, # JWT expiration time (up to 10 minutes)
    'iss': APP_ID     # Issuer, the GitHub App ID
}

# Generate the JWT
jwt_token = jwt.encode(payload, private_key_bytes, algorithm='RS256')

print(jwt_token)

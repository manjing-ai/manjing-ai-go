import os
import sys
from qcloud_cos import CosConfig
from qcloud_cos import CosS3Client

def upload_and_sign(secret_id, secret_key, region, bucket, local_path, remote_path):
    # Set Scheme to https
    config = CosConfig(Region=region, SecretId=secret_id, SecretKey=secret_key, Token=None, Scheme='https')
    client = CosS3Client(config)

    print(f"Uploading {local_path} to {remote_path}...", file=sys.stderr)
    try:
        response = client.upload_file(
            Bucket=bucket,
            LocalFilePath=local_path,
            Key=remote_path,
            EnableMD5=False
        )
    except Exception as e:
        print(f"Upload failed: {e}", file=sys.stderr)
        raise

    # Generate presigned URL (valid for 10 minutes)
    url = client.get_presigned_url(
        Method='GET',
        Bucket=bucket,
        Key=remote_path,
        Expired=600
    )
    return url

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python upload_cos.py <local_file> <remote_path>", file=sys.stderr)
        sys.exit(1)

    local_file = sys.argv[1]
    remote_path = sys.argv[2]
    
    secret_id = os.environ.get("TENCENT_SECRET_ID")
    secret_key = os.environ.get("TENCENT_SECRET_KEY")
    region = os.environ.get("COS_REGION")
    bucket = os.environ.get("COS_BUCKET")

    if not all([secret_id, secret_key, region, bucket]):
        print("Error: Missing env vars: TENCENT_SECRET_ID, TENCENT_SECRET_KEY, COS_REGION, COS_BUCKET", file=sys.stderr)
        sys.exit(1)

    try:
        url = upload_and_sign(secret_id, secret_key, region, bucket, local_file, remote_path)
        # Use the specific output format or just print URL to stdout
        # We will print to stdout so the shell script can capture it
        print(f"::set-output name=url::{url}")
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

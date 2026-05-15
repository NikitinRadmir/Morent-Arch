#!/bin/sh
set -e
ALIAS="local"
URL="http://minio:9000"

until mc alias set "$ALIAS" "$URL" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}"; do
  echo "minio-init: waiting for MinIO..."
  sleep 1
done

mc mb "${ALIAS}/${MINIO_BUCKET}" --ignore-existing
mc anonymous set download "${ALIAS}/${MINIO_BUCKET}"

# Политика жизненного цикла: удаление объектов через 30 дней после создания (окно хранения).
# Разные версии mc используют `ilm rule add` или `ilm add`.
if mc ilm rule add "${ALIAS}/${MINIO_BUCKET}" --id "morent-expire-30d" --expire-days 30 2>/dev/null; then
  echo "minio-init: ILM rule added (rule add)"
elif mc ilm add "${ALIAS}/${MINIO_BUCKET}" --expire-days 30 2>/dev/null; then
  echo "minio-init: ILM rule added (ilm add)"
else
  echo "minio-init: warning: could not set ILM expire-days=30 (check mc version)"
fi

echo "minio-init: done"

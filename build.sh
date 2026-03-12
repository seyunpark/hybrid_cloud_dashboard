#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

echo "=== Building Vineyard ==="
echo ""

# 테스트 컨테이너와 충돌하지 않도록 기존 대시보드 컨테이너 정리
if docker compose ps -q 2>/dev/null | grep -q .; then
  echo "Stopping existing dashboard containers..."
  docker compose down
  echo ""
fi

# 빌드
echo "=== Building Docker images ==="
docker compose build --no-cache
echo ""

# 실행
echo "=== Starting services ==="
docker compose up -d
echo ""

# 헬스체크 대기
echo "Waiting for backend to be healthy..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "Backend is ready!"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "Backend failed to start. Check logs: docker compose logs backend"
    exit 1
  fi
  sleep 2
done

echo ""
echo "=== Running ==="
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo ""
echo "Logs:    docker compose logs -f"
echo "Stop:    docker compose down"

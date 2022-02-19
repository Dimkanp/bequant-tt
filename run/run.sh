docker build -f ../Dockerfile -t cryptocompare:latest ../

docker-compose up -d && docker-compose logs -f

docker build -t musicbot-build .
docker run --name musicbot-build -v $(pwd):/home/musicbot musicbot-build gb build
docker rm musicbot-build
docker rmi musicbot-build

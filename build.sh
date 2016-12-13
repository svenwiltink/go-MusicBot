docker build -t musicbot-build .
docker run --name musicbot-build -v $(pwd):/home/musicbot musicbot-build sh -c "gb build; gb vendor restore"
docker rm musicbot-build

FROM golang:latest
WORKDIR /app
COPY ./ /app
RUN apt-get upgrade && apt-get install -y libssl1.1 wget dpkg 
RUN wget https://github.com/neo4j-drivers/seabolt/releases/download/v1.7.4/seabolt-1.7.4-Linux-ubuntu-18.04.deb
RUN dpkg -i seabolt-1.7.4-Linux-ubuntu-18.04.deb \
    && rm seabolt-1.7.4-Linux-ubuntu-18.04.deb
RUN go mod download
RUN go get github.com/githubnemo/CompileDaemon
ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main
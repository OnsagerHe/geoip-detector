FROM golang:1.23.4

RUN apt-get update && apt-get install -y \
    docker.io \
    wget \
    gnupg \
    && wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list' \
    && apt-get update && apt-get install -y \
    google-chrome-stable \
    && apt-get clean

RUN ln -s /usr/bin/google-chrome /usr/bin/chrome

WORKDIR /app

COPY go.* ./
RUN go mod tidy
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["/app/main", "-endpoint=https://onsager.net", "-prod=false", "-server=false"]

FROM node:lts-alpine as febuild
WORKDIR /work

COPY fe .

RUN yarn && yarn build


FROM golang:alpine as serverbuild

WORKDIR /work

COPY server .
COPY --from=febuild /work/dist /work/http_server/dist

RUN apk update && apk add git
RUN go build -ldflags "-X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail main.go


FROM alpine

WORKDIR /work

# 设置时区
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    &&rm -rf /var/cache/apk/* /tmp/* /var/tmp/* $HOME/.cache


COPY --from=serverbuild /work/pmail .
COPY server/config/dkim ./config/dkim/
COPY server/config/config.json ./config/

CMD /work/pmail
FROM node:lts-alpine as febuild
WORKDIR /work

COPY fe .

RUN yarn && yarn build


FROM golang:alpine as serverbuild
ARG VERSION
WORKDIR /work
COPY . .
COPY --from=febuild /work/dist /work/server/http_server/dist
RUN apk update && apk add git
RUN cd /work/server && go build -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.goVersion=$(go version)' -X 'main.gitHash=$(git show -s --format=%H)' -X 'main.buildTime=$(TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" -o pmail main.go
RUN cd /work/server/hooks/telegram_push && go build -ldflags "-s -w" -o output/telegram_push telegram_push.go
RUN cd /work/server/hooks/wechat_push && go build -ldflags "-s -w" -o output/wechat_push wechat_push.go
RUN cd /work/server/hooks/spam_block && go build -ldflags "-s -w" -o output/spam_block spam_block.go


FROM alpine

WORKDIR /work

# 设置时区
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    &&rm -rf /var/cache/apk/* /tmp/* /var/tmp/* $HOME/.cache


COPY --from=serverbuild /work/server/pmail .
COPY --from=serverbuild /work/server/hooks/telegram_push/output/* ./plugins/
COPY --from=serverbuild /work/server/hooks/wechat_push/output/* ./plugins/
COPY --from=serverbuild /work/server/hooks/spam_block/output/* ./plugins/

EXPOSE 25 80 110 443 465 995

CMD /work/pmail

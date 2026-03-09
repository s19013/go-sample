# デプロイ用コンテナに含めるバイナリを作成するコンテナ

# FROM golang:1.18.2-bullseye as deploy-helper
FROM golang:1.24.0-bullseye as deploy-helper

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# go build	: ビルド
# -o app	: 出力をappに
# -trimpath	: ローカルパス削除
# -ldflags "-w -s"	: デバッグ情報とシンボル削除
RUN go build -trimpath -ldflags "-w -s" -o app

# ----
# デプロイ用のコンテナ

FROM debian:bullseye-slim as deploy

RUN apt-get update

COPY --from=deploy-builder /app/app .

CMD [ "./app" ]

# ---
# ローカル開発環境で利用するホットロード環境
FROM golang:1.24 as dev
WORKDIR /app

RUN go install github.com/air-verse/air@latest
CMD [ "air" ]
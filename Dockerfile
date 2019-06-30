FROM golang:1.12 as build-env
ENV GO111MODULE=on

WORKDIR /app
ADD . /app

RUN go mod download
RUN go build
RUN ls /app

FROM gcr.io/distroless/base
COPY --from=build-env /app/miniCommerce /miniCommerce
COPY --from=build-env /app/sku_DJx1hCHoxDAAtE.pdf /sku_DJx1hCHoxDAAtE.pdf
COPY --from=build-env /app/sku_DWJE6B88Ih3Wgg.pdf /sku_DWJE6B88Ih3Wgg.pdf

CMD ["/miniCommerce"]

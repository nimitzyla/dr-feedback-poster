FROM --platform=linux/amd64 golang:alpine as builder
RUN apk update && apk add --no-cache git
RUN mkdir -p $GOPATH/src/wallmount-job
ADD . $GOPATH/src/wallmount-job
WORKDIR $GOPATH/src/wallmount-job
RUN go get -d -v
RUN go build -o wallmount-job .
# Stage 2
FROM --platform=linux/amd64 alpine
RUN mkdir /app
COPY --from=builder /go/src/wallmount-job/wallmount-job /app/
COPY --from=builder /go/src/wallmount-job/.env /app/
COPY --from=builder /go/src/wallmount-job/pn-bold.ttf /app/
COPY --from=builder /go/src/wallmount-job/pn.ttf /app/
COPY --from=builder /go/src/wallmount-job/template.jpg /app/
COPY --from=builder /go/src/wallmount-job/output/ /app/
COPY --from=builder /go/src/wallmount-job/images/ /app/
#ENV TZ=Asia/Kolkata
ENV ZONEINFO=/zoneinfo.zip
ARG APP_VERSION
ARG APP_NAME
ARG MODULE_NAME
ENV APP_VERSION=$APP_VERSION
ENV MODULE_NAME = $MODULE_NAME
ENV APP_NAME = $APP_NAME
EXPOSE 8080
WORKDIR /app
CMD ["./wallmount-job"]


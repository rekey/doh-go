FROM  alpine
LABEL maintainer="Rekey <rekey@me.com>"

WORKDIR /app/
ENV TZ=Asia/Shanghai
ADD ../../dist/web /app/

VOLUME /app/store
EXPOSE 54413

CMD ["./web"]

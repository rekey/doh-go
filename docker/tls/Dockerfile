FROM  alpine
LABEL maintainer="Rekey <rekey@me.com>"

WORKDIR /app/
ENV TZ=Asia/Shanghai
ADD ../../dist/web /app/
ENV ssl=dns.local

VOLUME /app/store
EXPOSE 54413

CMD ["./web"]

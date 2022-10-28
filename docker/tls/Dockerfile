FROM  alpine
LABEL maintainer="Rekey <rekey@me.com>"

WORKDIR /app/
ENV TZ=Asia/Shanghai
ADD ./tls /app/
ENV ssl=dns.local

VOLUME /app/store

CMD ["./tls"]

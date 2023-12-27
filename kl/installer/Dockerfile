FROM nginx

WORKDIR /app
RUN curl -s https://i.jpillora.com/installer | bash

ENTRYPOINT [ "./installer", "--force-user=kloudlite", "--force-repo=kl", "-u","kloudlite" ]

FROM mongo:6.0.5

WORKDIR /jsons
ENV MONGO_SCRIPT=script_runner.js
ENV SCRIPTS_FOLDER=/jsons
ENV MONGO_SCRIPTS_PORT=27017
ENV MONGO_SCRIPTS_HOST=localhost

COPY mongo_indexes/*.js ./
COPY mongo_indexes/docker-entrypoint.sh docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "/jsons/docker-entrypoint.sh"]


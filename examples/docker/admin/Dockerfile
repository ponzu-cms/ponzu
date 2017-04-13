FROM krismeister/ponzu

ENV PONZU_SHARE /ponzu-share

# This is where you want to mount your project:
ENV PROJECT_FOLDER /go/src/project

VOLUME $PONZU_SHARE
RUN mkdir $PROJECT_FOLDER

WORKDIR $PROJECT_FOLDER
EXPOSE 8080

# its better to run this in compose
# CMD [ "bash" , "/go/src/project/start_admin.sh start" ]
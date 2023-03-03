#
# Dockerfile for client tester
#LABEL author="Ziad Khalaf"
#THIS IS A STATEMENT THAT INDICATES THE LATEST COMMIT IS COPIED
FROM python:3

WORKDIR /home
RUN pip install requests

COPY . .
CMD [ "python", "/home/scenario1.py" ]
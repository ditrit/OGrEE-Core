#
# Dockerfile for client tester
#LABEL author="Ziad Khalaf"
FROM python:3

WORKDIR /home
RUN pip install requests

COPY . .
CMD [ "python", "/home/scenario1.py" ]
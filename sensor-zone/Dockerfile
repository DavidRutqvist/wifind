FROM node:alpine

# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Install app dependencies
COPY package.json /usr/src/app/
RUN npm install

# Bundle app source
COPY . /usr/src/app

# Build app
RUN npm run grunt

# Remove source from container
RUN rm -rf /src

EXPOSE 3000
CMD [ "npm", "start" ]

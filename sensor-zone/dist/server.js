"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const bodyParser = require("body-parser");
const express = require("express");
const logger = require("morgan");
const errorHandler = require("errorhandler");
const log = require("winston");
const units_1 = require("./routes/units");
const database_1 = require("./database");
class Server {
    static bootstrap() {
        return new Server();
    }
    constructor() {
        this.app = express();
        this.config();
        this.routes();
    }
    config() {
        this.app.use(logger("dev"));
        this.app.use(bodyParser.json());
        this.app.use(function (err, req, res, next) {
            err.status = 404;
            next(err);
        });
        this.app.use(errorHandler());
        database_1.Database.initialize(process.env.MONGO_CONNECTION_STRING);
    }
    routes() {
        const router = express.Router();
        units_1.UnitsRoute.create(router);
        this.app.use("/", router);
    }
    shutdown(callback) {
        log.info("Shutting down server");
        callback();
    }
}
exports.Server = Server;

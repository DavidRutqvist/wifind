"use strict";
import * as bodyParser from "body-parser";
import * as express from "express";
import * as logger from "morgan";
import * as errorHandler from "errorhandler";
import * as Rx from "rxjs/Rx";
import * as log from "winston";
import { UnitsRoute } from "./routes/units";
import { Database } from "./database";

/**
 * The server.
 *
 * @class Server
 */
export class Server {
  public app: express.Application;

  /**
   * Bootstrap the application.
   *
   * @class Server
   * @method bootstrap
   * @static
   * @return {Server} Returns the newly created server
   */
  public static bootstrap(): Server {
    return new Server();
  }

  /**
   * Constructor.
   *
   * @class Server
   * @constructor
   */
  constructor() {
    // create expressjs application
    this.app = express();

    // configure application
    this.config();

    // add routes
    this.routes();
  }

  /**
   * Configure application
   *
   * @class Server
   * @method config
   */
  public config(): void {
    // use logger middlware
    this.app.use(logger("dev"));

    // use json form parser middlware
    this.app.use(bodyParser.json());

    this.app.use(function(err: any, req: express.Request, res: express.Response, next: express.NextFunction): void {
      err.status = 404;
      next(err);
    });

    // error handling
    this.app.use(errorHandler());

    // set up database
    Database.initialize(process.env.MONGO_CONNECTION_STRING);
  }

  /**
   * Create router
   *
   * @class Server
   * @method api
   */
  public routes(): void {
    const router: express.Router = express.Router();

    // create routes
    UnitsRoute.create(router);

    // use router middleware
    this.app.use("/", router);
  }

  public shutdown(callback: () => void): void {
    log.info("Shutting down server");

    // perform any shutdown logic here

    callback();
  }
}

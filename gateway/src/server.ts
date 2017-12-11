import * as bodyParser from "body-parser";
import * as express from "express";
import * as logger from "morgan";
import errorHandler = require("errorhandler");
import * as Rx from "rxjs/Rx";
import { IndexRoute } from "./routes/index";
import { ZonesRoute } from "./routes/zones";
import { SensorsRoute } from "./routes/sensors";
import { ServiceFactory } from "./services/service-factory";
import { ServiceDiscovery } from "./utils/service-discovery";
import { SocketHandler } from "./push/socket-handler";

/**
 * The server.
 *
 * @class Server
 */
export class Server {
  public app: express.Application;
  private serviceFactory: ServiceFactory;
  private socketServer: SocketIO.Server;

  /**
   * Bootstrap the application.
   *
   * @class Server
   * @method bootstrap
   * @static
   * @return {Server} Returns the newly created server for this app.
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
  
  public setSocketServer(socketServer: SocketIO.Server): void {
    this.socketServer = socketServer;
    this.socketServer.on("connection", socket => SocketHandler.onConnection(socket));
  }

  /**
   * Configure application
   *
   * @class Server
   * @method config
   */
  public config(): void {
    // set up logging
    this.app.use(logger("dev"));

    // set up json body parsing
    this.app.use(bodyParser.json());
    this.app.use(bodyParser.urlencoded({ extended: false }));

    // allow CORS
    this.app.use(function(req: express.Request, res: express.Response, next: express.NextFunction): void {
      res.header("Access-Control-Allow-Origin", "*");
      res.header("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS");
      res.header("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With");

      // intercept OPTIONS method (CORS Preflight)
      if ("OPTIONS" === req.method) {
        res.send(200);
      } else {
        next();
      }
    });

    // catch 404 and forward to error handler
    this.app.use(function(err: any, req: express.Request, res: express.Response, next: express.NextFunction): void {
        err.status = 404;
        next(err);
    });

    // error handling
    this.app.use(errorHandler());

    // set up service discovery
    const discovery: ServiceDiscovery = new ServiceDiscovery(process.env.CONSUL_ADDR ? process.env.CONSUL_ADDR : "localhost:8500");

    // set up service factory
    this.serviceFactory = new ServiceFactory(discovery);

    // init socket handler
    SocketHandler.initialize();
  }

  /**
   * Create router
   *
   * @class Server
   * @method api
   */
  public routes(): void {
    const router: express.Router = express.Router();

    IndexRoute.create(router);
    SensorsRoute.create(router, this.serviceFactory);
    ZonesRoute.create(router, this.serviceFactory);

    // use router middleware for /api-routes this to "reserve" the /-path for documentation or some ordinary webpage
    this.app.use("/api", router);
  }

  public shutdown(callback: () => void): void {
    console.log("Shutting down server");
    callback();
  }
}

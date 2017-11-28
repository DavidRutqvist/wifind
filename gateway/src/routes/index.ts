"use strict";
import { NextFunction, Request, Response, Router } from "express";
import log from "winston";

/**
 * Defines the index route
 *
 * @class IndexRoute
 */
export class IndexRoute {
  /**
   * Create the routes.
   *
   * @class IndexRoute
   * @method create
   * @static
   */
  public static create(router: Router): void {
    log.info("[IndexRoute::create] Creating index route.");

    router.get("/", (req: Request, res: Response, next: NextFunction) => {
      res.json({
        success: true,
        message: "Up and running"
      });
    });
  }
}

"use strict";
import { NextFunction, Request, Response, Router } from "express";
import * as log from "winston";

/**
 * Defines the route for zone endpoints
 *
 * @class ZonesRoute
 */
export class ZonesRoute {
  /**
   * Create the routes.
   *
   * @class ZonesRoute
   * @method create
   * @static
   */
  public static create(router: Router): void {
    log.info("[ZonesRoute::create] Creating zones route.");

    router.get("/zones", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute().getZones(req, res, next);
    });
  }

  public getZones(req: Request, res: Response, next: NextFunction): void {
    res.status(500).json({
      success: false,
      message: "Not yet implemented"
    });
  }
}

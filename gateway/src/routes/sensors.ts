"use strict";
import { NextFunction, Request, Response, Router } from "express";

/**
 * Defines the route for sensor endpoints
 *
 * @class SensorsRoute
 */
export class SensorsRoute {
  /**
   * Create the routes.
   *
   * @class SensorsRoute
   * @method create
   * @static
   */
  public static create(router: Router): void {
    console.log("[SensorsRoute::create] Creating sensors route.");

    router.get("/users", (req: Request, res: Response, next: NextFunction) => {
      res.json({users: [{name: "David"}, {name: "Test Testsson"}]});
    });
  }
}

"use strict";
import { NextFunction, Request, Response, Router } from "express";

/**
 * Defines the route for users endpoints
 *
 * @class IndexRoute
 */
export class UsersRoute {
  /**
   * Create the routes.
   *
   * @class UsersRoute
   * @method create
   * @static
   */
  public static create(router: Router): void {
    console.log("[UsersRoute::create] Creating users route.");

    router.get("/users", (req: Request, res: Response, next: NextFunction) => {
      res.json({users: [{name: "David"}, {name: "Test Testsson"}]});
    });
  }
}

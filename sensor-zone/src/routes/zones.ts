"use strict";
import { NextFunction, Request, Response, Router } from "express";
import * as log from "winston";
import { Database } from "../database";
import { ZoneNotFoundError } from '../errors/zone-not-found-error';

export class ZonesRoute {
  public static create(router: Router): void {
    log.info("[ZonesRoute::create] Creating zones route.");

    router.get("/zones/:zoneId", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute().getZoneSensors(req, res, next);
    });
  }

  constructor() { }

  public getZoneSensors(req: Request, res: Response, next: NextFunction): void {
    Database.getInstance().getZoneSensors(req.params.zoneId)
      .subscribe(
        sensors => res.json({
          success: true,
          sensors: sensors
        }),
        err => this.logAndReturn(res, err)
      );
  }

  private logAndReturn(res: Response, err: Error): void {
    if (err instanceof ZoneNotFoundError) {
      res.status(404).json({
        success: false,
        message: "The zone could not be found"
      });
    } else {
      log.error("Error in API", err);
      res.status(500).json({
        success: false,
        message: "Something went wrong"
      });
    }
  }
}

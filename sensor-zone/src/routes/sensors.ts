"use strict";
import { SensorNotFoundError } from '../errors/sensor-not-found-error';
import { NextFunction, Request, Response, Router } from "express";
import * as log from "winston";
import { Database } from "../database";

export class SensorsRoute {
  public static create(router: Router): void {
    log.info("[SensorsRoute::create] Creating sensors route.");

    router.get("/sensors/:sensorid", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute().getSensor(req, res, next);
    });

    router.get("/sensors/:sensorid/zones", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute().getZones(req, res, next);
    });

    router.post("/sensors/:sensorid/zones", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute().addSensor(req, res, next);
    });
  }

  constructor() { }

  public getSensor(req: Request, res: Response, next: NextFunction): void {
    //verify the id parameter exists
    Database.getInstance().getSensor(req.params.sensorid)
    .subscribe(
      sensor => res.json({
        success: true,
        sensor: sensor
      }),
      err => this.logAndReturn(res, err)
    );
  }
  
  public getZones(req: Request, res: Response, next: NextFunction): void {
    Database.getInstance().getLocations(req.params.sensorid)
      .subscribe(
        locations => res.json({
          success: true,
          locations: locations
        }),
        err => this.logAndReturn(res, err)
      );
  }

  public addSensor(req: Request, res: Response, next: NextFunction): void {
    if(! req.body.zoneId){
      res.status(400).json({
        success: false,
        message: "Missing zoneId in body"
      })
      return;
    }

    if(! req.body.from){
      res.status(400).json({
        success: false,
        message: "Missing from in body"
      })
      return;
    }

    const fromTimestamp: number = req.body.from;
    const from: Date = new Date(fromTimestamp * 1000); // times 1000 for ms -> seconds
    
    let to: Date = undefined;
    if (req.body.to) {
      const toTimestamp: number = req.body.to;
      to = new Date(toTimestamp * 1000);
    }

    Database.getInstance().updateLastIfNull(req.params.sensorid, from)
      .flatMap(() => Database.getInstance().addSensorLocation(req.params.sensorid, req.body.zoneId, from, to))
      .subscribe(
        location => res.json({
          success: true,
          location: location
        }),
        err => this.logAndReturn(res, err)
      );
  }

  private logAndReturn(res: Response, err: Error): void {
    if (err instanceof SensorNotFoundError) {
      res.status(404).json({
        success: false,
        message: "The sensor could not be found"
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

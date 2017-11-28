"use strict";
import { Message } from '_debugger';
import { SensorNotFoundError } from '../errors/sensor-not-found-error';
import { NextFunction, Request, Response, Router } from "express";
import * as log from 'winston';
import { Database } from "../database";

export class UnitsRoute {
  public static create(router: Router): void {
    log.info("[UnitsRoute::create] Creating units route.");

    router.get("/sensors/:sensorid", (req: Request, res: Response, next: NextFunction) => {
      new UnitsRoute().getSensor(req, res, next);
    });

    router.get("/sensors/:sensorid/zones", (req: Request, res: Response, next: NextFunction) => {
      new UnitsRoute().getZones(req, res, next);
    });
    router.get("/sensors/zones/:zoneId", (req: Request, res: Response, next: NextFunction) => {
      new UnitsRoute().getZoneSensors(req, res, next);
    });

    router.post("/sensors", (req: Request, res: Response, next: NextFunction) => {
      new UnitsRoute().addSensor(req, res, next);
    });

    router.put("/sensors/:sensorid", (req: Request, res: Response, next: NextFunction) => {
      new UnitsRoute().updateSensor(req, res, next);
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

    if(! req.body.sensorId){
      res.status(400).json({
        success: false,
        message: "Missing sensorId in body"
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

    Database.getInstance().updateLastIfNull(req.body.sensorId, from)
      .flatMap(() => Database.getInstance().addSensorLocation(req.body.sensorId, req.body.zoneId, from, to))
      .subscribe(
        location => res.json({
          success: true,
          location: location
        }),
        err => this.logAndReturn(res, err)
      );
  }

  public updateSensor(req: Request, res: Response, next: NextFunction): void {
    res.status(400).json({
      success: false,
      message: "Not yet implemented"
    });
  }

  public deleteSensor(req: Request, res: Response, next: NextFunction): void {
    res.status(400).json({
      success: false,
      message: "Not yet implemented"
    });
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

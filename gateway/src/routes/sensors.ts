"use strict";
import { NextFunction, Request, Response, Router } from "express";
import { ServiceFactory } from "../services/service-factory";
import { AxiosHelper } from "../utils/axios-helper";
import * as log from "winston";

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
  public static create(router: Router, serviceFactory: ServiceFactory): void {
    log.info("[SensorsRoute::create] Creating sensors route.");

    router.get("/sensors", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute(serviceFactory).getSensors(req, res, next);
    });

    router.get("/sensors/:id", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute(serviceFactory).getSensor(req, res, next);
    });

    router.get("/sensors/:id/occupancy", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute(serviceFactory).getSensorOccupancy(req, res, next);
    });

    router.get("/sensors/:id/observations", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute(serviceFactory).getSensorObservations(req, res, next);
    });

    router.get("/sensors/:id/locations", (req: Request, res: Response, next: NextFunction) => {
      new SensorsRoute(serviceFactory).getSensorLocations(req, res, next);
    });
  }

  constructor(private readonly serviceFactory: ServiceFactory) { }

  public getSensors(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getDatastoreService()
      .flatMap(svc => svc.getSensors())
      .subscribe(
          sensors => res.json({
              success: true,
              sensors: sensors
          }),
          err => AxiosHelper.handleError(err, res));
  }

  public getSensor(req: Request, res: Response, next: NextFunction): void {
    res.status(500).json({
      success: false,
      message: "Not yet implemented"
    });
  }

  public getSensorLocations(req: Request, res: Response, next: NextFunction): void {
    res.status(500).json({
      success: false,
      message: "Not yet implemented"
    });
  }

  public getSensorOccupancy(req: Request, res: Response, next: NextFunction): void {
    res.status(500).json({
      success: false,
      message: "Not yet implemented"
    });
  }

  public getSensorObservations(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getDatastoreService()
      .flatMap(svc => svc.getSensor(req.params.id))
      .subscribe(
          sensor => res.json({
              success: true,
              sensor: sensor
          }),
          err => AxiosHelper.handleError(err, res));
  }

}

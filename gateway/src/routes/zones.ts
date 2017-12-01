"use strict";
import { NextFunction, Request, Response, Router } from "express";
import * as log from "winston";
import { ServiceFactory } from "../services/service-factory";
import { AxiosHelper } from "../utils/axios-helper";

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
  public static create(router: Router, serviceFactory: ServiceFactory): void {
    log.info("[ZonesRoute::create] Creating zones route.");

    router.get("/zones", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getZones(req, res, next);
    });

    router.post("/zones", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).createZone(req, res, next);
    });

    router.get("/zones/:id", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getZone(req, res, next);
    });

    router.get("/zones/:id/sensors", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getSensors(req, res, next);
    });

    router.get("/zones/:id/children", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getZoneChildren(req, res, next);
    });

    router.get("/zones/:id", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getZone(req, res, next);
    });

    router.get("/zones/:id/occupation", (req: Request, res: Response, next: NextFunction) => {
      new ZonesRoute(serviceFactory).getOccupation(req, res, next);
    });
  }

  constructor(private readonly serviceFactory: ServiceFactory) { }

  public getZones(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getZonesService()
      .flatMap(svc => svc.getZones())
      .subscribe(
        zones => res.json({
          success: true,
          zones: zones
        }),
        err => AxiosHelper.handleError(err, res));
  }

  public getZone(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getZonesService()
      .flatMap(svc => svc.getZone(req.params.id))
      .subscribe(
        zone => res.json({
          success: true,
          zone: zone
        }),
        err => AxiosHelper.handleError(err, res));
  }

  public getZoneChildren(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getZonesService()
      .flatMap(svc => svc.getChildren(req.params.id))
      .subscribe(
        children => res.json({
          success: true,
          children: children
        }),
        err => AxiosHelper.handleError(err, res));
  }

  public createZone(req: Request, res: Response, next: NextFunction): void {
    if (!req.body.name) {
      res.status(400).json({
        success: false,
        message: "Missing mandatory name"
      });
      return;
    }

    if (!req.body.location) {
      res.status(400).json({
        success: false,
        message: "Missing mandatory location"
      });
      return;
    }

    if (req.body.location.length !== 2) {
      res.status(400).json({
        success: false,
        message: "Location parameter must be exactly two items long"
      });
      return;
    }

    const createObservable = this.serviceFactory.getZonesService()
      .flatMap(svc => svc.createZone(req.body.name, req.body.location, req.body.parent))
      .subscribe(
        result => {
          if (result) {
            res.json({
              success: true,
              message: "Zone successfully created"
            });
          } else {
            res.status(500).json({
              success: false,
              message: "Could not create zone"
            });
          }
        },
        err => AxiosHelper.handleError(err, res)
      );
  }

  public getSensors(req: Request, res: Response, next: NextFunction): void {
    this.serviceFactory.getSensorLocationService()
      .flatMap(svc => svc.getSensors(req.params.id))
      .subscribe(
        sensors => res.json({
          success: true,
          sensors: sensors
        }),
        err => AxiosHelper.handleError(err, res));
  }

  public getOccupation(req: Request, res: Response, next: NextFunction): void {
    res.status(500).json({
      success: false,
      message: "Not yet implemented"
    });
  }
}

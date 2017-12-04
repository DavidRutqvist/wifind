"use strict";
import { ServiceDiscovery } from "../utils/service-discovery";
import { ZonesService } from "./zones/zones.service";
import { DatastoreService } from "./datastore/datastore.service";
import { ServiceTypes } from "../utils/service-types";
import * as Rx from "rxjs/Rx";
import { SensorLocationService } from "./sensor-location/sensor-location.service";

export class ServiceFactory {
  constructor(private readonly services: ServiceDiscovery) { }

  public getZonesService(): Rx.Observable<ZonesService> {
    return this.services.getServiceUri(ServiceTypes.Zones, "http")
      .map(uri => new ZonesService("http://" + uri));
  }

  public getDatastoreService(): Rx.Observable<DatastoreService> {
    return this.services.getServiceUri(ServiceTypes.Datastore, "http")
      .map(uri => new DatastoreService("http://" + uri));
  }

  public getSensorLocationService(): Rx.Observable<SensorLocationService> {
    return this.services.getServiceUri(ServiceTypes.SensorLocation, "http")
      .map(uri => new SensorLocationService("http://" + uri));
  }
}

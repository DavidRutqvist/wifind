"use strict";
import { ServiceDiscovery } from "../utils/service-discovery";
import { ZonesService } from "./zones/zones.service";
import { ServiceTypes } from "../utils/service-types";
import * as Rx from "rxjs/Rx";

export class ServiceFactory {
  constructor(private readonly services: ServiceDiscovery) { }

  public getZonesService(): Rx.Observable<ZonesService> {
    return this.services.getServiceUri(ServiceTypes.Zones, "http")
      .map(uri => new ZonesService("http://" + uri));
  }
}
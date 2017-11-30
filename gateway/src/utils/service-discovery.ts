"use strict";
import * as Consul from "consul";
import * as Rx from "rxjs/Rx";
import * as log from "winston";

export class ServiceDiscovery {
  private readonly consul: Consul.Consul;

  constructor(consulAddress: string) {
    const addressParts = consulAddress.split(":");
    this.consul = Consul({
      host: addressParts[0],
      port: addressParts.length > 1 ? addressParts[1] : consulAddress,
      promisify: true
    });
  }

  public getServiceUri(service: string, tag?: string): Rx.Observable<string> {
    return Rx.Observable.fromPromise(this.consul.catalog.service.nodes({
      service: service,
      tag: tag
    }))
    .map(x => <any[]>x)
    .flatMap(x => x)
    .first()
    .map(instance => this.toUri(instance));
  }

  private toUri(instance: any): string {
    return instance.ServiceAddress + ":" + instance.ServicePort;
  }
}
import { Injectable } from "@angular/core";
import * as Rx from "rxjs/Rx";
import { ConfigService } from "app/shared/services/config/config.service";
import { Zone } from "app/shared/services/zones/zone";

@Injectable()
export class ZonesService {

  constructor(private config: ConfigService) { }

  public getTopZones(): Rx.Observable<Zone[]> {
    return Rx.Observable.fromPromise(this.config.getApiClient().get("/zones"))
      .map(res => res.data)
      .map(data => data.zones)
      .flatMap(zones => this.getArrayOrThrow<Zone>(zones))
      .catch(err => Rx.Observable.of([]));
  }

  public getZoneChildren(zoneId: string): Rx.Observable<Zone[]> {
    return Rx.Observable.fromPromise(this.config.getApiClient().get("/zones/" + zoneId + "/children"))
      .map(res => res.data)
      .map(data => data.children)
      .flatMap(children => this.getArrayOrThrow<Zone>(children))
      .catch(err => Rx.Observable.of([]));
  }

  public getZone(zoneId: string): Rx.Observable<Zone> {
    return Rx.Observable.fromPromise(this.config.getApiClient().get("/zones/" + zoneId))
      .map(res => res.data)
      .map(data => data.zone);
  }

  public getAllZones(): Rx.Observable<Zone[]> {
    return this.getTopZones()
      .flatMap(x => x)
      .flatMap(zone => this.getFlattenChildren(zone))
      .flatMap(x => x)
      .toArray();
  }

  public addSensorToZone(zoneId: string, sensorId: string): Rx.Observable<void> {
    return Rx.Observable.fromPromise(this.config.getApiClient().post("/zones/" + zoneId + "/sensors", {
      sensorId: sensorId,
      from: Math.round(new Date().getTime() / 1000)
    }))
      .map(res => res.data)
      .flatMap(res => this.throwOnUnsuccessful(res));
  }

  private getFlattenChildren(zone: Zone): Rx.Observable<Zone[]> {
    return this.getZoneChildren(zone.id)
      .flatMap(x => x)
      .flatMap(child => this.getFlattenChildren(child))
      .flatMap(x => x)
      .toArray()
      .map(flatChildren => {
        flatChildren.unshift(zone);
        return flatChildren;
      });
  }

  private getArrayOrThrow<T>(arr: T[]): Rx.Observable<T[]> {
    if (arr) {
      return Rx.Observable.of(arr);
    } else {
      return Rx.Observable.throw(new Error("Array cannot be null or undefined"));
    }
  }

  private throwOnUnsuccessful(result: any): Rx.Observable<void> {
    if (result && result.success === true) {
      return Rx.Observable.of(this.noop());
    } else {
      if (result.message) {
        return Rx.Observable.throw(new Error(result.message));
      } else {
        return Rx.Observable.throw(new Error("Something went wrong"));
      }
    }
  }

  private noop(): void { }
}

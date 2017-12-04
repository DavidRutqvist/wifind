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

  private getArrayOrThrow<T>(arr: T[]): Rx.Observable<T[]> {
    if (arr) {
      return Rx.Observable.of(arr);
    } else {
      return Rx.Observable.throw(new Error("Array cannot be null or undefined"));
    }
  }
}

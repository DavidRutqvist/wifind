import { Injectable } from "@angular/core";
import * as Rx from "rxjs/Rx";
import { ConfigService } from "app/shared/services/config/config.service";

@Injectable()
export class SensorsService {

  constructor(private config: ConfigService) { }

  public getSensors(): Rx.Observable<string[]> {
    return Rx.Observable.fromPromise(this.config.getApiClient().get("/sensors"))
      .map(res => res.data)
      .map(data => data.sensors)
      .flatMap(sensors => this.getArrayOrThrow<string>(sensors))
      .catch(err => Rx.Observable.of([]));
  }

  private getArrayOrThrow<T>(arr: T[]): Rx.Observable<T[]> {
    if (arr) {
      return Rx.Observable.of(arr);
    } else {
      return Rx.Observable.throw(new Error("Array cannot be null or undefined"));
    }
  }
}

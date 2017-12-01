"use strict";
import * as Rx from "rxjs/Rx";
import * as log from "winston";
import axios, { AxiosInstance } from "axios";
import { SensorLocation } from "./sensor-location";

export class SensorLocationService {
  constructor(private readonly serviceUri: string) { }

  public getLocations(sensorId: string): Rx.Observable<SensorLocation[]> {
    return Rx.Observable.fromPromise(this.getAxios().get("/sensors/" + sensorId))
      .map(res => res.data);
  }
  
  public getSensors(zoneId: string): Rx.Observable<SensorLocation[]> {
    return Rx.Observable.fromPromise(this.getAxios().get("/zones/" + zoneId))
      .map(res => res.data)
      .catch(err => this.catch404(err));
  }

  public addSensorToZone(zoneId: string, sensorId: string, from: number, to?: number): Rx.Observable<boolean> {
    return Rx.Observable.fromPromise(this.getAxios().post("/sensors/" + sensorId + "/zones", {
      zoneId: zoneId,
      from: from,
      to: to
    }))
      .map(res => res.data.success);
  }

  private getAxios(): AxiosInstance {
    return axios.create({
      baseURL: this.serviceUri,
      timeout: 1000
    });
  }

  private catch404(err: Error): Rx.Observable<SensorLocation[]> {
    if ((<any>err).response && (<any>err).response.status === 404) {
      return Rx.Observable.of([]);
    } else {
      return Rx.Observable.throw(err);
    }
  }
}

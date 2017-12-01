"use strict";
import * as Rx from "rxjs/Rx";
import log from "winston";
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
      .map(res => res.data);
  }

  private getAxios(): AxiosInstance {
    return axios.create({
      baseURL: this.serviceUri,
      timeout: 1000
    });
  }
}

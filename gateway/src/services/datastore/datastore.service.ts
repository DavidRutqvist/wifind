"use strict";
import { DeviceObservation } from "./datastore";
import * as Rx from "rxjs/Rx";
import log from "winston";
import axios, { AxiosInstance } from "axios";

export class DatastoreService {
  constructor(private readonly serviceUri: string) { }

  public getSensors(): Rx.Observable<DeviceObservation[]> {
    return Rx.Observable.fromPromise(this.getAxios().get("/sensor"))
      .map(res => res.data);
  }

  public getSensor(id: string): Rx.Observable<DeviceObservation> {
    return Rx.Observable.fromPromise(this.getAxios().get("/sensor/" + id))
      .map(res => res.data);
  }

  private getAxios(): AxiosInstance {
    return axios.create({
      baseURL: this.serviceUri,
      timeout: 1000
    });
  }
}

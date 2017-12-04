"use strict";
import { Zone } from "./zone";
import * as Rx from "rxjs/Rx";
import log from "winston";
import axios, { AxiosInstance } from "axios";

export class ZonesService {
  constructor(private readonly serviceUri: string) { }

  public getZones(): Rx.Observable<Zone[]> {
    return Rx.Observable.fromPromise(this.getAxios().get("/zones"))
      .map(res => res.data);
  }

  public getZone(id: string): Rx.Observable<Zone> {
    return Rx.Observable.fromPromise(this.getAxios().get("/zones/" + id))
      .map(res => res.data);
  }
  
  public getChildren(id: string): Rx.Observable<string[]> {
    return Rx.Observable.fromPromise(this.getAxios().get("/zones/" + id + "/children"))
      .map(res => res.data);
  }

  public createZone(name: string, location: number[], parent?: string): Rx.Observable<boolean> {
    return Rx.Observable.fromPromise(this.getAxios().post("/zones", {
      name: name,
      location: location,
      parent: parent
    }))
      .map(res => res.data.success);
  }

  private getAxios(): AxiosInstance {
    return axios.create({
      baseURL: this.serviceUri,
      timeout: 1000
    });
  }
}

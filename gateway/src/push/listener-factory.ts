"use strict";
import { AmqpTopicListener } from "./amqp-topic-listener";
import * as Rx from "rxjs/Rx";
import { ServiceTypes } from "../utils/service-types";
import { ServiceDiscovery } from "../utils/service-discovery";
import { OccupancyListener } from "./occupancy/occupancy-listener";

export class ListenerFactory {
  public static getListeners(serviceDiscovery: ServiceDiscovery, exchange: string): Rx.Observable<AmqpTopicListener[]> {
    const amqp = serviceDiscovery.getServiceUri("rabbit", "amqp")
    .map(uri => "amqp://" + uri)
    .share();

    const occupancyListener = amqp.map(connectionString => new OccupancyListener(connectionString, exchange))

    return Rx.Observable.forkJoin([occupancyListener]);
  }
}
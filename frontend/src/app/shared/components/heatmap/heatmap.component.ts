import { Component, OnInit, Input, ElementRef, ViewChild } from "@angular/core";
import { Zone } from "app/shared/services/zones/zone";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { ZoneOccupancy } from "app/shared/services/zones/zone-occupancy";
import * as Rx from "rxjs/Rx";
import { google } from "google-maps";
declare var google: any;

@Component({
  selector: "app-heatmap",
  templateUrl: "./heatmap.component.html",
  styleUrls: ["./heatmap.component.scss"]
})
export class HeatmapComponent implements OnInit {
  @Input() zones: Zone[];
  private occupancyMap: { [id: string]: number} = {};
  map: google.maps.Map;
  heatmap: google.maps.visualization.HeatmapLayer;
  @ViewChild("map") mapElement: ElementRef;

  constructor(
    private readonly zonesSvc: ZonesService,
    private elementRef: ElementRef) { }

  ngOnInit() {
    const zoneIds: string[] = [];
    for (let i = 0; i < this.zones.length; i++) {
      zoneIds.push(this.zones[i].id);
      this.occupancyMap[this.zones[i].id] = 0;

      this.zonesSvc.getRealtimeOccupancy([this.zones[i].id])
        .subscribe(zoneOccupancy => this.updateOccupancy(zoneOccupancy));
    }

    /*this.zonesSvc.getRealtimeOccupancy(zoneIds)
      .subscribe(zoneOccupancy => console.log(zoneOccupancy));
    //*/

    this.initMap();
  }

  private updateOccupancy(occupancy: ZoneOccupancy): void {
    this.occupancyMap[occupancy.zone] = occupancy.occupancy;
    this.updateHeatmap();
  }

  initMap(): void {
    this.map = new google.maps.Map(this.mapElement.nativeElement, {
      center: new google.maps.LatLng(65.618035, 22.138819),
      zoom: 15,
      mapTypeId: google.maps.MapTypeId.ROADMAP
    });

    this.heatmap = new google.maps.visualization.HeatmapLayer({
      data: this.getPoints(),
      map: this.map,
      radius: 75,
      maxIntensity: 200
    });
  }

  updateHeatmap(): void {
    if (this.heatmap) {
      this.heatmap.setData(this.getPoints());
    }
  }

  private getPoints(): google.maps.visualization.WeightedLocation[] {
    const points: any[] = [];

    for (let i = 0; i < this.zones.length; i++) {
        points.push({
          location: new google.maps.LatLng(this.zones[i].location[0], this.zones[i].location[1]),
          weight: this.occupancyMap[this.zones[i].id]
        });
    }

    return points;
  }
}

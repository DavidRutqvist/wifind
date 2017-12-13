import { Component, OnInit, Input } from "@angular/core";
import { Zone } from "app/shared/services/zones/zone";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { ZoneOccupancy } from "app/shared/services/zones/zone-occupancy";

@Component({
  selector: "app-heatmap",
  templateUrl: "./heatmap.component.html",
  styleUrls: ["./heatmap.component.scss"]
})
export class HeatmapComponent implements OnInit {
  @Input() zones: Zone[];
  private occupancyMap: { [id: string]: number} = {};

  constructor(private readonly zonesSvc: ZonesService) { }

  ngOnInit() {
    const zoneIds: string[] = [];
    for (let i = 0; i < this.zones.length; i++) {
      zoneIds.push(this.zones[i].id);
      this.occupancyMap[this.zones[i].id] = 0;
    }

    this.zonesSvc.getRealtimeOccupancy(zoneIds)
      .subscribe(zoneOccupancy => this.updateOccupancy(zoneOccupancy));
  }

  private updateOccupancy(occupancy: ZoneOccupancy): void {
    this.occupancyMap[occupancy.zoneId] = occupancy.occupancy;
  }
}

import { Component, OnInit, Input } from "@angular/core";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { Zone } from "app/shared/services/zones/zone";

@Component({
  selector: "app-zone-badge",
  templateUrl: "./zone-badge.component.html",
  styleUrls: ["./zone-badge.component.scss"]
})
export class ZoneBadgeComponent implements OnInit {
  @Input() zone: Zone;
  occupancy = 0;
  occupancyHistory: any[] = [];

  constructor(private readonly zoneSvc: ZonesService) { }

  ngOnInit() {
    this.zoneSvc.getRealtimeOccupancy([this.zone.id])
      .subscribe(zoneOccupancy => this.occupancyChanged(zoneOccupancy.occupancy));
  }

  private occupancyChanged(occupancy: number): void {
    this.occupancy = occupancy;
  }
}

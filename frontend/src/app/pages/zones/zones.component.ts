import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Zone } from 'app/shared/services/zones/zone';
import { ZonesService } from 'app/shared/services/zones/zones.service';

@Component({
  selector: 'app-zones',
  templateUrl: './zones.component.html',
  styleUrls: ['./zones.component.scss']
})
export class ZonesComponent implements OnInit, OnDestroy {
  id: string;
  private sub: any;
  zone: Zone;

  constructor(private route: ActivatedRoute, private readonly zoneSvc: ZonesService) { }

  ngOnInit() {
    this.sub = this.route.params.subscribe(params => {
      this.id = params["id"];
      this.initZone();
   });
  }

  ngOnDestroy() {
    this.sub.unsubscribe();
  }

  initZone() {
    this.zoneSvc.getZone(this.id)
      .subscribe(zone => this.zone = zone);
  }
}

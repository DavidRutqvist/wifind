import {
  Component,
  ViewEncapsulation,
  ViewChild,
  OnInit,
  HostListener,
  ElementRef
} from "@angular/core";
import { GlobalState } from "../../app.state";
import { ConfigService } from "../../shared/services/config/config.service";
import { MdSidenav } from "@angular/material";
import { ZonesService } from "app/shared/services/zones/zones.service";
import { Zone } from "app/shared/services/zones/zone";

@Component({
  selector: ".content_inner_wrapper",
  templateUrl: "./dashboards.component.html",
  styleUrls: ["./dashboards.component.scss"],
  encapsulation: ViewEncapsulation.Emulated
})
export class DashboardsComponent implements OnInit {
  //Header Title
  title: string = "Dashboard";
  zones: Zone[];

  public doughnutChartLabels: string[] = [
    "Writing Code",
    "Problem Solving",
    "Debugging",
    "Designing"
  ];

  public doughnutChartData: number[] = [350, 450, 100, 220];
  public doughnutChartType: string = "doughnut";
  public doughnutChartColors: Array<any> = [
    {
      backgroundColor: ["#796AEE", "#28BEBD", "#2196F3", "#EC407A"]
    }
  ];

  // events
  public chartClicked(e: any): void {
    console.log(e);
  }

  public chartHovered(e: any): void {
    console.log(e);
  }

  constructor(
    public config: ConfigService,
    private _elementRef: ElementRef,
    private _state: GlobalState,
    private readonly zoneSvc: ZonesService
  ) {

  }

  ngOnInit() {
    this.zoneSvc.getAllZones()
      .subscribe(zones => this.zones = zones);
  }


}

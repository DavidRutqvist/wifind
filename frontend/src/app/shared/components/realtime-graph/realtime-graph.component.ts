import { Component, OnInit, Input } from "@angular/core";
import * as Highcharts from "highcharts";
import { ZonesService } from "app/shared/services/zones/zones.service";
import * as Rx from "rxjs/Rx";

@Component({
  selector: "app-realtime-graph",
  templateUrl: "./realtime-graph.component.html",
  styleUrls: ["./realtime-graph.component.scss"]
})
export class RealtimeGraphComponent implements OnInit {
  @Input() zoneId: string;
  @Input() height = 300;
  occupancyData: any[] = [];
  chart: any;

  options = {
    title: "",
    subtitle: "",
    chart: {
      zoomType: "xy",
      height: this.height
    },
    credits: {
      enabled: false
    },
    xAxis: {
        type: "datetime"
    },
    series: [{
      type: "line",
      name: "Occupancy",
      visible: true,
      data: [],
      animation: true,
      marker: {
          enabled: false
      }
    }]
  };

  constructor(private readonly zoneSvc: ZonesService) { }

  ngOnInit() {
    this.options.chart.height = this.height;
    this.zoneSvc.getRealtimeOccupancy([this.zoneId])
      .subscribe(occupancy => this.updateOccupancy(occupancy));
    Rx.Observable.timer(0, 30000) // check every 30 seconds
      .subscribe(() => this.removeOldSamples());
  }

  private updateOccupancy(occupancy: number): void {
    this.occupancyData.push([new Date().getTime(), occupancy]);
    this.chart.series[0].setData(this.occupancyData);
  }

  private removeOldSamples(): void {
    for (let i = 0; i < this.occupancyData.length; i++) {
      if ((new Date().getTime() - this.occupancyData[i][0]) >= 3600 * 1000) { // remove older than 1 hour
        this.occupancyData.splice(i, 1);
      } else {
        return;
      }
    }
  }

  private saveInstance(chart: any) {
    this.chart = chart;
  }
}

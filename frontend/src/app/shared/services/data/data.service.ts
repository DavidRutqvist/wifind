import { Http, Headers, Response, Jsonp, RequestOptions } from "@angular/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs/Observable";

import "rxjs/Rx";
import "rxjs/add/operator/map";
import "rxjs/add/operator/toPromise";
import "rxjs/add/operator/catch";
@Injectable()
export class DataService {
  constructor(private http: Http) { }

  getProfile = (): Observable<Response> => {
    return this.http
      .get("/assets/data/profile/profile.data.json")
      .map(res => res.json());
  };
  getContactsCardDemo = (): Observable<Response> => {
    return this.http
      .get("/assets/data/contacts/contacts-card.data.json")
      .map(res => res.json());
  };
  getMailDemo = (): Observable<Response> => {
    return this.http
      .get("/assets/data/mail/mail.json")
      .map(res => res.json());
  };
  getListCardDemo = (): Observable<Response> => {
    return this.http
      .get("/assets/data/list/ListCardData.json")
      .map(res => res.json());
  };
	getChatContacts = (): Observable<Response> => {
		return this.http
			.get("/assets/data/chat/messages.data.json")
			.map(res => res.json());
	};
}

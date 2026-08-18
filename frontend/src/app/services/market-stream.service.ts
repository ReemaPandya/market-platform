import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { Quote } from './trading.service';

@Injectable({
  providedIn: 'root',
})
export class MarketStreamService {
  streamQuote(symbol: string): Observable<Quote> {
    return new Observable<Quote>((observer) => {
      const wsBaseUrl = environment.apiUrl.replace(/^http/, 'ws');

      const socket = new WebSocket(
        `${wsBaseUrl}/market/ws/${symbol.trim().toUpperCase()}`
      );

      socket.onmessage = (event) => {
        try {
          const quote: Quote = JSON.parse(event.data);
          observer.next(quote);
        } catch (error) {
          observer.error(error);
        }
      };

      socket.onerror = (error) => {
        observer.error(error);
      };

      socket.onclose = () => {
        observer.complete();
      };

      return () => {
        socket.close();
      };
    });
  }
}
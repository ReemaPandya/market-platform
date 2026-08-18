import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { AuthService } from './auth.service';

export interface Position {
  symbol: string;
  exchange: string;
  quantity: number;
  average_price: number;
  current_price: number;
  invested_value: number;
  current_value: number;
  unrealized_pnl: number;
}

export interface Order {
  id: number;
  symbol: string;
  exchange: string;
  side: string;
  quantity: number;
  executed_price: number;
  status: string;
  created_at: string;
}

export interface CreateOrderRequest {
  symbol: string;
  exchange: string;
  side: 'BUY' | 'SELL';
  quantity: number;
}

export interface Quote {
  symbol: string;
  exchange: string;
  ltp: number;
  source: string;
}

@Injectable({
  providedIn: 'root',
})
export class TradingService {
  private apiUrl = environment.apiUrl;

  constructor(
    private http: HttpClient,
    private authService: AuthService,
  ) {}

  private authHeaders(): HttpHeaders {
    const token = this.authService.getToken();

    return new HttpHeaders({
      Authorization: `Bearer ${token}`,
    });
  }

  getPortfolio(): Observable<{ positions: Position[] }> {
    return this.http.get<{ positions: Position[] }>(
      `${this.apiUrl}/trading/portfolio`,
      {
        headers: this.authHeaders(),
      }
    );
  }

  getOrders(): Observable<{ orders: Order[] }> {
    return this.http.get<{ orders: Order[] }>(
      `${this.apiUrl}/trading/orders`,
      {
        headers: this.authHeaders(),
      }
    );
  }

  createOrder(
    order: CreateOrderRequest
  ): Observable<any> {
    return this.http.post(
      `${this.apiUrl}/trading/orders`,
      order,
      {
        headers: this.authHeaders(),
      }
    );
  }

  getQuote(symbol: string): Observable<Quote> {
    return this.http.get<Quote>(
      `${this.apiUrl}/market/quote/${symbol.toUpperCase()}`
    );
  }
}
import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

import { AuthService } from '../../services/auth.service';
import {
  Order,
  Position,
  Quote,
  TradingService,
} from '../../services/trading.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css',
})
export class DashboardComponent implements OnInit {
  positions: Position[] = [];
  orders: Order[] = [];

  loading = true;
  error = '';

  symbol = 'INFY';
  quantity = 1;
  side: 'BUY' | 'SELL' = 'BUY';

  quote: Quote | null = null;
  orderMessage = '';
  orderError = '';
  submittingOrder = false;

  constructor(
    private authService: AuthService,
    private tradingService: TradingService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    this.loadDashboard();
    this.loadQuote();
  }

  loadDashboard(): void {
    this.loading = true;
    this.error = '';

    this.tradingService.getPortfolio().subscribe({
      next: (response) => {
        this.positions = response.positions;

        this.tradingService.getOrders().subscribe({
          next: (orderResponse) => {
            this.orders = orderResponse.orders;
            this.loading = false;
          },
          error: () => {
            this.error = 'Unable to load order history.';
            this.loading = false;
          },
        });
      },

      error: (err) => {
        console.error(err);

        if (err.status === 401) {
          this.authService.logout();
          this.router.navigate(['/login']);
          return;
        }

        this.error = 'Unable to load portfolio.';
        this.loading = false;
      },
    });
  }

  loadQuote(): void {
    const normalizedSymbol = this.symbol.trim().toUpperCase();

    if (!normalizedSymbol) {
      return;
    }

    this.symbol = normalizedSymbol;

    this.tradingService.getQuote(normalizedSymbol).subscribe({
      next: (quote) => {
        this.quote = quote;
      },
      error: () => {
        this.quote = null;
      },
    });
  }

  submitOrder(): void {
    this.orderMessage = '';
    this.orderError = '';

    if (!this.symbol.trim()) {
      this.orderError = 'Enter a stock symbol.';
      return;
    }

    if (this.quantity <= 0) {
      this.orderError = 'Quantity must be greater than zero.';
      return;
    }

    this.submittingOrder = true;

    this.tradingService.createOrder({
      symbol: this.symbol.trim().toUpperCase(),
      exchange: 'NSE',
      side: this.side,
      quantity: this.quantity,
    }).subscribe({
      next: (response) => {
        this.orderMessage =
          `${response.side} order filled at ₹${response.executed_price}`;

        this.submittingOrder = false;

        this.loadDashboard();
        this.loadQuote();
      },

      error: (err) => {
        this.orderError =
          err?.error?.error || 'Unable to place order.';

        this.submittingOrder = false;
      },
    });
  }

  totalInvested(): number {
    return this.positions.reduce(
      (total, position) => total + position.invested_value,
      0
    );
  }

  totalCurrentValue(): number {
    return this.positions.reduce(
      (total, position) => total + position.current_value,
      0
    );
  }

  totalPnL(): number {
    return this.positions.reduce(
      (total, position) => total + position.unrealized_pnl,
      0
    );
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
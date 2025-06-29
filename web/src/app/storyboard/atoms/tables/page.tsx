"use client";

import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/atoms";
import { Badge } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

const invoices = [
  {
    invoice: "INV001",
    paymentStatus: "Paid",
    totalAmount: "$250.00",
    paymentMethod: "Credit Card",
  },
  {
    invoice: "INV002",
    paymentStatus: "Pending",
    totalAmount: "$150.00",
    paymentMethod: "PayPal",
  },
  {
    invoice: "INV003",
    paymentStatus: "Unpaid",
    totalAmount: "$350.00",
    paymentMethod: "Bank Transfer",
  },
  {
    invoice: "INV004",
    paymentStatus: "Paid",
    totalAmount: "$450.00",
    paymentMethod: "Credit Card",
  },
  {
    invoice: "INV005",
    paymentStatus: "Paid",
    totalAmount: "$550.00",
    paymentMethod: "PayPal",
  },
];

export default function TablesPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Table Component
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Responsive table components for displaying structured data.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Table"
        description="Simple table with header and body"
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead className="text-right">Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>John Doe</TableCell>
              <TableCell>john@example.com</TableCell>
              <TableCell>Admin</TableCell>
              <TableCell className="text-right">Active</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Jane Smith</TableCell>
              <TableCell>jane@example.com</TableCell>
              <TableCell>User</TableCell>
              <TableCell className="text-right">Active</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Bob Johnson</TableCell>
              <TableCell>bob@example.com</TableCell>
              <TableCell>User</TableCell>
              <TableCell className="text-right">Inactive</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </ComponentShowcase>

      <ComponentShowcase
        title="Table with Caption"
        description="Table with a descriptive caption"
      >
        <Table>
          <TableCaption>A list of your recent invoices.</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Invoice</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Method</TableHead>
              <TableHead className="text-right">Amount</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {invoices.map((invoice) => (
              <TableRow key={invoice.invoice}>
                <TableCell className="font-medium">{invoice.invoice}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      invoice.paymentStatus === "Paid"
                        ? "success"
                        : invoice.paymentStatus === "Pending"
                        ? "warning"
                        : "destructive"
                    }
                  >
                    {invoice.paymentStatus}
                  </Badge>
                </TableCell>
                <TableCell>{invoice.paymentMethod}</TableCell>
                <TableCell className="text-right">{invoice.totalAmount}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </ComponentShowcase>

      <ComponentShowcase
        title="Table with Footer"
        description="Table with footer for totals"
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Invoice</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Method</TableHead>
              <TableHead className="text-right">Amount</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {invoices.slice(0, 3).map((invoice) => (
              <TableRow key={invoice.invoice}>
                <TableCell className="font-medium">{invoice.invoice}</TableCell>
                <TableCell>{invoice.paymentStatus}</TableCell>
                <TableCell>{invoice.paymentMethod}</TableCell>
                <TableCell className="text-right">{invoice.totalAmount}</TableCell>
              </TableRow>
            ))}
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={3}>Total</TableCell>
              <TableCell className="text-right">$750.00</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      </ComponentShowcase>

      <ComponentShowcase
        title="Striped Table"
        description="Table with alternating row colors"
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Product</TableHead>
              <TableHead>Category</TableHead>
              <TableHead className="text-right">Price</TableHead>
              <TableHead className="text-right">Stock</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[
              { product: "Laptop", category: "Electronics", price: "$999", stock: "12" },
              { product: "Mouse", category: "Accessories", price: "$29", stock: "143" },
              { product: "Keyboard", category: "Accessories", price: "$79", stock: "87" },
              { product: "Monitor", category: "Electronics", price: "$299", stock: "23" },
              { product: "Headphones", category: "Audio", price: "$149", stock: "56" },
            ].map((item, index) => (
              <TableRow key={item.product} className={index % 2 === 0 ? "bg-muted/50" : ""}>
                <TableCell>{item.product}</TableCell>
                <TableCell>{item.category}</TableCell>
                <TableCell className="text-right">{item.price}</TableCell>
                <TableCell className="text-right">{item.stock}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </ComponentShowcase>

      <ComponentShowcase
        title="Responsive Table"
        description="Table that scrolls horizontally on small screens"
      >
        <div className="w-full overflow-auto">
          <Table className="min-w-[600px]">
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Customer</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Phone</TableHead>
                <TableHead>Country</TableHead>
                <TableHead className="text-right">Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>001</TableCell>
                <TableCell>2024-01-15</TableCell>
                <TableCell>Alice Johnson</TableCell>
                <TableCell>alice@example.com</TableCell>
                <TableCell>+1-555-0123</TableCell>
                <TableCell>USA</TableCell>
                <TableCell className="text-right">$1,234.56</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>002</TableCell>
                <TableCell>2024-01-16</TableCell>
                <TableCell>Bob Williams</TableCell>
                <TableCell>bob@example.com</TableCell>
                <TableCell>+1-555-0124</TableCell>
                <TableCell>Canada</TableCell>
                <TableCell className="text-right">$2,345.67</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </ComponentShowcase>
    </div>
  );
}
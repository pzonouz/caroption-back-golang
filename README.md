How to use restore.sh
chmod +x restore.sh 
./restore.sh /path/to/caroption_go-2025-01-03.tar.gz


-- ==========================================
-- 1. ENUMS (Data Types)
-- ==========================================

-- انواع حساب‌ها (دارایی، بدهی، حقوق صاحبان سهام، درآمد، هزینه)
CREATE TYPE account_type_enum AS ENUM ('Asset', 'Liability', 'Equity', 'Revenue', 'Expense');

-- وضعیت اسناد حسابداری (موقت، تایید شده، قطعی/بسته شده)
CREATE TYPE voucher_status_enum AS ENUM ('Draft', 'Approved', 'Locked');

-- نوع تراکنش خزانه‌داری (دریافت، پرداخت، انتقال بانک‌به‌بانک)
CREATE TYPE transaction_type_enum AS ENUM ('Receipt', 'Payment', 'Transfer');

-- روش‌های پرداخت/دریافت (نقد، بانک/کارتخوان، چک)
CREATE TYPE payment_method_enum AS ENUM ('Cash', 'Bank', 'Check');

-- نوع چک (دریافتنی، پرداختنی)
CREATE TYPE check_type_enum AS ENUM ('Receivable', 'Payable');

-- وضعیت چک (ثبت‌شده، در جریان وصول، پاس‌شده، برگشتی، عودت‌داده‌شده)
CREATE TYPE check_status_enum AS ENUM ('Registered', 'In_Bank', 'Cleared', 'Bounced', 'Returned');

-- نوع فاکتور (فروش، خرید، برگشت از فروش، برگشت از خرید)
CREATE TYPE invoice_type_enum AS ENUM ('Sales', 'Purchase', 'Sales_Return', 'Purchase_Return');

-- وضعیت فاکتور (پیش‌نویس، صادرشده، تسویه‌شده، باطل‌شده)
CREATE TYPE invoice_status_enum AS ENUM ('Draft', 'Issued', 'Paid', 'Cancelled');

-- ==========================================
-- 2. ACCOUNTING TABLES
-- ==========================================

-- جدول درخت حساب‌ها (کدینگ حسابداری)
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,                    -- کد حساب (مثلا 1010)
    name VARCHAR(100) NOT NULL,                          -- نام حساب (مثلا بانک ملت یا هزینه برق)
    type account_type_enum NOT NULL,                     -- ماهیت حساب
    parent_id INT REFERENCES accounts(id),               -- شناسه حساب پدر (برای ساختار درختی)
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول هدر اسناد حسابداری (Vouchers)
CREATE TABLE vouchers (
    id SERIAL PRIMARY KEY,
    voucher_number VARCHAR(50) UNIQUE NOT NULL,          -- شماره سند
    date DATE NOT NULL,                                  -- تاریخ سند
    description TEXT,                                    -- شرح کل سند
    status voucher_status_enum DEFAULT 'Draft',
    created_by INT,                                      -- شناسه کاربری سیستم
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول اقلام سند حسابداری (ردیف‌های بدهکار و بستانکار)
CREATE TABLE voucher_lines (
    id SERIAL PRIMARY KEY,
    voucher_id INT NOT NULL REFERENCES vouchers(id),
    account_id INT NOT NULL REFERENCES accounts(id),
    debit NUMERIC(18, 2) DEFAULT 0.00,                   -- مبلغ بدهکار
    credit NUMERIC(18, 2) DEFAULT 0.00,                  -- مبلغ بستانکار
    description VARCHAR(255),                            -- شرح ردیف
    
    CONSTRAINT check_positive_amounts CHECK (debit >= 0 AND credit >= 0)
);

-- ==========================================
-- 3. TREASURY TABLES
-- ==========================================

-- جدول مدیریت چک‌ها
CREATE TABLE checks (
    id SERIAL PRIMARY KEY,
    type check_type_enum NOT NULL,                       -- دریافتی/پرداختی
    check_number VARCHAR(50) NOT NULL,                   -- شماره سریال چک
    bank_name VARCHAR(100) NOT NULL,                     -- نام بانک
    due_date DATE NOT NULL,                              -- تاریخ سررسید
    amount NUMERIC(18, 2) NOT NULL,
    issuer_name VARCHAR(150),                            -- صادرکننده/گیرنده
    status check_status_enum DEFAULT 'Registered',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول هدر تراکنش‌های مالی (فرم رسید دریافت / پرداخت)
CREATE TABLE financial_transactions (
    id SERIAL PRIMARY KEY,
    type transaction_type_enum NOT NULL,                 -- نوع فرم (Receipt, Payment, Transfer)
    date DATE NOT NULL,
    total_amount NUMERIC(18, 2) NOT NULL,                -- جمع کل مبلغ فرم
    contact_id INT,                                      -- لینک به مشتری/تامین‌کننده
    offset_account_id INT REFERENCES accounts(id),       -- حساب مقابل (مثلا هزینه حقوق یا درآمد فروش)
    voucher_id INT REFERENCES vouchers(id),              -- لینک به سند حسابداری صادر شده (اتوماتیک)
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول اقلام تراکنش (ترکیب نقد، کارتخوان و چک در یک فرم)
CREATE TABLE transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INT NOT NULL REFERENCES financial_transactions(id),
    payment_method payment_method_enum NOT NULL,         -- نقد، بانک، چک
    amount NUMERIC(18, 2) NOT NULL,
    source_dest_account_id INT REFERENCES accounts(id),  -- صندوق یا بانک مرتبط
    reference_number VARCHAR(100),                       -- شماره فیش / پیگیری
    check_id INT REFERENCES checks(id),                  -- در صورت چک بودن
    
    CONSTRAINT check_payment_method_logic CHECK (
        (payment_method = 'Check' AND check_id IS NOT NULL) OR 
        (payment_method != 'Check' AND check_id IS NULL)
    )
);

-- ==========================================
-- 4. INVOICING TABLES
-- ==========================================

-- جدول اشخاص (مشتریان، تامین‌کنندگان، پرسنل)
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    contact_type VARCHAR(50),
    account_id INT REFERENCES accounts(id),          -- لینک به حساب تفصیلی شخص در کدینگ
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول کالاها و خدمات
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    price NUMERIC(18, 2) NOT NULL,
    income_account_id INT REFERENCES accounts(id),   -- حساب درآمدی کالا
    expense_account_id INT REFERENCES accounts(id)   -- حساب هزینه‌ای کالا
);

-- جدول هدر فاکتورها
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    type invoice_type_enum NOT NULL,                 -- فروش / خرید
    contact_id INT NOT NULL REFERENCES contacts(id), -- لینک به مشتری
    date DATE NOT NULL,
    
    subtotal NUMERIC(18, 2) DEFAULT 0.00,            -- جمع مبلغ پایه
    total_discount NUMERIC(18, 2) DEFAULT 0.00,      -- جمع تخفیفات
    total_tax NUMERIC(18, 2) DEFAULT 0.00,           -- جمع مالیات
    grand_total NUMERIC(18, 2) DEFAULT 0.00,         -- مبلغ نهایی فاکتور
    
    status invoice_status_enum DEFAULT 'Draft',
    voucher_id INT REFERENCES vouchers(id),          -- سند حسابداری اتوماتیک فاکتور
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- جدول اقلام فاکتور
CREATE TABLE invoice_items (
    id SERIAL PRIMARY KEY,
    invoice_id INT NOT NULL REFERENCES invoices(id),
    product_id INT NOT NULL REFERENCES products(id),
    quantity NUMERIC(10, 2) NOT NULL,
    unit_price NUMERIC(18, 2) NOT NULL,
    discount NUMERIC(18, 2) DEFAULT 0.00,
    tax NUMERIC(18, 2) DEFAULT 0.00,
    line_total NUMERIC(18, 2) NOT NULL               -- = (qty * unit_price) - discount + tax
);
# ۵. مثال عملی از روند کار سیستم (Workflow Example)

فرض کنید می‌خواهیم یک **فاکتور فروش** برای شرکت "آلفا" به مبلغ پایه $10,000,000$ تومان و مالیات $900,000$ تومان صادر کنیم، و سپس مشتری این مبلغ را به صورت ترکیبی ($5,000,000$ نقد و $5,900,000$ چک) پرداخت کند.

## مرحله اول: صدور فاکتور فروش

1. کاربر فاکتور را در جدول `invoices` با مبلغ نهایی $10,900,000$ ثبت می‌کند.
2. سیستم بک‌اند به طور خودکار سند حسابداری زیر را در `voucher_lines` (ردیف‌های سند) ثبت می‌کند:

| شرح (Account) | مبلغ بدهکار (Debit) | مبلغ بستانکار (Credit) |
| :--- | :--- | :--- |
| حساب مشتری (شرکت آلفا) | $10,900,000$ | $0$ |
| درآمد فروش | $0$ | $10,000,000$ |
| مالیات بر ارزش افزوده | $0$ | $900,000$ |
| **جمع کل** | **$10,900,000$** | **$10,900,000$** |

> **نکته:** همان‌طور که می‌بینید، شرط تراز بودن سند حسابداری ($\sum Debit = \sum Credit$) کاملاً رعایت شده است.

---

## مرحله دوم: تسویه فاکتور (دریافت وجه)

مشتری بدهی خود را پرداخت می‌کند. این عملیات در زیرسیستم خزانه‌داری ثبت می‌شود:

1. در جدول `financial_transactions` یک فرم `Receipt` (رسید دریافت) ثبت می‌شود که در آن `offset_account_id` (حساب مقابل) برابر با **حساب مشتری (شرکت آلفا)** است.
2. در جدول `transaction_items` دو ردیف برای این فرم ثبت می‌شود:
   * یک ردیف از نوع `Cash` (نقد) به مبلغ $5,000,000$.
   * یک ردیف از نوع `Check` (چک) به مبلغ $5,900,000$ (با لینک به اطلاعات چک در جدول `checks`).
3. سیستم بک‌اند پس از ذخیره این فرم، بلافاصله **سند تسویه** زیر را به صورت خودکار در هسته حسابداری (`vouchers` و `voucher_lines`) صادر می‌کند:

| شرح (Account) | مبلغ بدهکار (Debit) | مبلغ بستانکار (Credit) |
| :--- | :--- | :--- |
| صندوق مرکزی (ورود پول نقد) | $5,000,000$ | $0$ |
| اسناد دریافتنی (ورود چک به صندوق) | $5,900,000$ | $0$ |
| حساب مشتری (شرکت آلفا) | $0$ | $10,900,000$ |
| **جمع کل** | **$10,900,000$** | **$10,900,000$** |

> **نکته:** در این سند نیز شرط تراز بودن ($\sum Debit = \sum Credit$) برقرار است.


---

## نتیجه‌گیری منطقی سیستم

با ثبت این دو سند در سیستم حسابداری دوطرفه:
* در **سند اول** (فاکتور)، حساب "شرکت آلفا" مبلغ $10,900,000$ **بدهکار** شد.
* در **سند دوم** (دریافت وجه)، حساب "شرکت آلفا" مبلغ $10,900,000$ **بستانکار** شد.

در نتیجه، مانده حساب "شرکت آلفا" در دفاتر حسابداری دقیقاً **صفر** می‌شود ($10,900,000 - 10,900,000 = 0$). با صفر شدن این مانده، سیستم متوجه می‌شود که فاکتور مربوطه کاملاً تسویه شده و وضعیت آن به `Paid` تغییر می‌کند.

# ۶. توصیه‌های فنی و بهترین تجربیات (Best Practices)

برای پیاده‌سازی اصولی این پایگاه داده و جلوگیری از بروز خطاهای رایج در سیستم‌های مالی، رعایت نکات زیر در سطح دیتابیس و بک‌اند (Back-end) الزامی است:

## ۱. دقت محاسباتی (Data Types)
در سیستم‌های مالی **هرگز** از نوع داده‌های `FLOAT`، `REAL` یا `DOUBLE` برای ذخیره مبالغ استفاده نکنید. این نوع داده‌ها در محاسبات ممیز شناور دچار خطای گردکردن (Rounding Error) می‌شوند. 
* همیشه از نوع داده ثابت مانند `NUMERIC(18, 2)` یا `DECIMAL(18, 2)` استفاده کنید تا محاسباتی مانند $10.01 + 20.02$ دقیقاً برابر با $30.03$ شود.

## ۲. استفاده از تراکنش‌های پایگاه داده (ACID Transactions)
ثبت یک رویداد مالی معمولاً شامل درج داده در چندین جدول مختلف است (مثلاً ثبت همزمان در `invoices`، `vouchers` و `voucher_lines`). 
* تمام این دستورات `INSERT` باید درون یک تراکنش واحد (`BEGIN` و `COMMIT`) قرار بگیرند. 
* اگر در حین ثبت یکی از جداول خطایی رخ داد، کل عملیات باید `ROLLBACK` شود تا از ایجاد اسناد ناقص و ناتراز جلوگیری گردد.

## ۳. گزارش‌گیری سریع (Views & Materialized Views)
برای استخراج ترازنامه، تراز آزمایشی یا مانده حساب اشخاص، نیازی به ذخیره مقدار مانده (Balance) در جدول جداگانه نیست. 
* از یک `VIEW` در دیتابیس استفاده کنید که برای هر `account_id`، فرمول $\sum Debit - \sum Credit$ را به صورت زنده (Real-time) محاسبه کند. 
* برای سیستم‌های بسیار بزرگ با میلیون‌ها رکورد، استفاده از `MATERIALIZED VIEW` برای تهیه گزارش‌های ماهانه پیشنهاد می‌شود.


## ۵. قفل‌گذاری اسناد (Record Locking)
پس از پایان هر دوره مالی (مثلاً پایان ماه یا سال)، اسناد حسابداری آن دوره باید قفل شوند تا هیچ کاربری نتواند آن‌ها را ویرایش کند. 
* استفاده از وضعیت `Locked` در `voucher_status_enum` و کنترل آن در لایه بک‌اند (یا با استفاده از Database Triggers) از تغییرات غیرمجاز در دوره‌های بسته شده جلوگیری می‌کند.



# ۷. ماهیت حساب‌ها در فاکتورهای فروش و خرید

در حسابداری، هر حساب دارای یک «ماهیت» (ذاتی) است: **دارایی** و **هزینه** ماهیت بدهکار دارند (افزایش آن‌ها بدهکار ثبت می‌شود)؛ در حالی که **بدهی**، **سرمایه** و **درآمد** ماهیت بستانکار دارند (افزایش آن‌ها بستانکار ثبت می‌شود).

در ادامه ساختار استاندارد اسناد حسابداری برای فاکتورهای فروش و خرید آورده شده است:

## ۱. فاکتور فروش (Sales Invoice)

هنگام صدور فاکتور فروش، شما کالا یا خدماتی را ارائه می‌دهید، درآمد کسب می‌کنید و مشتری به شما بدهکار می‌شود.

**ماهیت حساب‌های درگیر:**
* **حساب مشتری (حساب‌های دریافتنی):** ماهیت **دارایی** دارد. چون طلب شما از مشتری افزایش می‌یابد، **بدهکار** می‌شود.
* **حساب درآمد فروش:** ماهیت **درآمد** دارد. چون درآمد شما افزایش یافته است، **بستانکار** می‌شود.
* **حساب مالیات بر ارزش افزوده (فروش):** ماهیت **بدهی** دارد (پولی است که از مشتری گرفته‌اید اما باید به دولت بدهید). بنابراین **بستانکار** می‌شود.

**نمونه سند حسابداری فاکتور فروش:**

| شرح حساب | گروه حساب (Nature) | مبالغ بدهکار (Debit) | مبالغ بستانکار (Credit) |
| :--- | :--- | :--- | :--- |
| حساب مشتری (مثلاً شرکت آلفا) | دارایی (Asset) | $10,900,000$ | $0$ |
| درآمد فروش کالا/خدمات | درآمد (Revenue) | $0$ | $10,000,000$ |
| مالیات و عوارض فروش | بدهی (Liability) | $0$ | $900,000$ |
| **جمع کل** | | **$10,900,000$** | **$10,900,000$** |

---

## ۲. فاکتور خرید (Purchase Invoice)

فاکتور خرید دقیقاً نقطه مقابل فاکتور فروش است. شما کالا یا خدماتی را دریافت می‌کنید، هزینه‌ای متحمل می‌شوید و به تامین‌کننده (فروشنده) بدهکار می‌شوید.

**ماهیت حساب‌های درگیر:**
* **حساب تامین‌کننده (حساب‌های پرداختنی):** ماهیت **بدهی** دارد. چون بدهی شما به فروشنده افزایش می‌یابد، **بستانکار** می‌شود.
* **حساب خرید / موجودی کالا:** بسته به روش حسابداری، ماهیت **هزینه** یا **دارایی** دارد. چون هزینه/موجودی شما افزایش یافته است، **بدهکار** می‌شود.
* **حساب مالیات بر ارزش افزوده (خرید):** ماهیت **دارایی** دارد (پولی است که به عنوان مالیات پرداخته‌اید و از بدهی مالیاتی شما به دولت کسر می‌شود - اعتبار مالیاتی). بنابراین **بدهکار** می‌شود.

**نمونه سند حسابداری فاکتور خرید:**

| شرح حساب | گروه حساب (Nature) | مبالغ بدهکار (Debit) | مبالغ بستانکار (Credit) |
| :--- | :--- | :--- | :--- |
| خرید کالا / هزینه خدمات | هزینه/دارایی (Expense/Asset)| $5,000,000$ | $0$ |
| مالیات بر ارزش افزوده خرید | دارایی (Asset) | $450,000$ | $0$ |
| حساب تامین‌کننده (مثلاً شرکت بتا) | بدهی (Liability) | $0$ | $5,450,000$ |
| **جمع کل** | | **$5,450,000$** | **$5,450,000$** |

---

## خلاصه قانون تراز ($\sum Debit = \sum Credit$)

* در **فاکتور فروش**: پولی که قرار است بگیرید (بدهکار) $=$ ارزش واقعی کالا (بستانکار) $+$ سهم دولت (بستانکار).
* در **فاکتور خرید**: پولی که باید بپردازید (بستانکار) $=$ ارزش واقعی کالا (بدهکار) $+$ سهم دولت که پیش‌پرداخت کرده‌اید (بدهکار).




(دارایی‌ها + هزینه‌ها = بدهی‌ها + سرمایه + درآمدها)



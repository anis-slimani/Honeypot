#!/usr/bin/env python3
"""
Script de conversion Markdown vers PDF
Convertit le cahier des charges en PDF professionnel
"""

import markdown
from weasyprint import HTML, CSS
from weasyprint.text.fonts import FontConfiguration
import sys

def convert_markdown_to_pdf(md_file, pdf_file):
    """Convertit un fichier Markdown en PDF avec style professionnel"""

    # Lire le fichier Markdown
    with open(md_file, 'r', encoding='utf-8') as f:
        md_content = f.read()

    # Convertir Markdown en HTML avec extensions
    html_content = markdown.markdown(
        md_content,
        extensions=['tables', 'fenced_code', 'nl2br', 'sane_lists']
    )

    # CSS professionnel pour le PDF
    css_style = """
    @page {
        size: A4;
        margin: 2.5cm 2cm;
        @top-right {
            content: "Page " counter(page) " / " counter(pages);
            font-size: 9pt;
            color: #666;
        }
    }

    body {
        font-family: 'DejaVu Sans', Arial, sans-serif;
        font-size: 10pt;
        line-height: 1.6;
        color: #333;
    }

    h1 {
        color: #2c3e50;
        font-size: 24pt;
        margin-top: 30pt;
        margin-bottom: 15pt;
        padding-bottom: 10pt;
        border-bottom: 3px solid #3498db;
        page-break-after: avoid;
    }

    h2 {
        color: #34495e;
        font-size: 18pt;
        margin-top: 25pt;
        margin-bottom: 12pt;
        padding-bottom: 5pt;
        border-bottom: 2px solid #95a5a6;
        page-break-after: avoid;
    }

    h3 {
        color: #555;
        font-size: 14pt;
        margin-top: 20pt;
        margin-bottom: 10pt;
        page-break-after: avoid;
    }

    h4 {
        color: #666;
        font-size: 12pt;
        margin-top: 15pt;
        margin-bottom: 8pt;
        page-break-after: avoid;
    }

    p {
        margin: 10pt 0;
        text-align: justify;
    }

    ul, ol {
        margin: 10pt 0;
        padding-left: 25pt;
    }

    li {
        margin: 5pt 0;
    }

    table {
        width: 100%;
        border-collapse: collapse;
        margin: 15pt 0;
        font-size: 9pt;
        page-break-inside: avoid;
    }

    th {
        background-color: #3498db;
        color: white;
        padding: 8pt;
        text-align: left;
        font-weight: bold;
        border: 1px solid #2980b9;
    }

    td {
        padding: 6pt 8pt;
        border: 1px solid #bdc3c7;
    }

    tr:nth-child(even) {
        background-color: #f8f9fa;
    }

    code {
        font-family: 'DejaVu Sans Mono', 'Courier New', monospace;
        background-color: #f4f4f4;
        padding: 2pt 4pt;
        border-radius: 3pt;
        font-size: 9pt;
        color: #c7254e;
    }

    pre {
        background-color: #f8f8f8;
        border: 1px solid #ddd;
        border-left: 3px solid #3498db;
        padding: 10pt;
        margin: 15pt 0;
        overflow-x: auto;
        page-break-inside: avoid;
        font-size: 8pt;
    }

    pre code {
        background-color: transparent;
        padding: 0;
        color: #333;
    }

    blockquote {
        border-left: 4px solid #3498db;
        padding-left: 15pt;
        margin: 15pt 0;
        color: #666;
        font-style: italic;
        background-color: #f9f9f9;
        padding: 10pt 15pt;
    }

    hr {
        border: none;
        border-top: 2px solid #bdc3c7;
        margin: 20pt 0;
    }

    strong {
        color: #2c3e50;
        font-weight: bold;
    }

    em {
        color: #555;
    }

    a {
        color: #3498db;
        text-decoration: none;
    }

    /* Éviter les coupures de page mal placées */
    h1, h2, h3, h4, h5, h6 {
        page-break-after: avoid;
    }

    table, figure, img {
        page-break-inside: avoid;
    }

    /* Style pour les badges de statut */
    .status-badge {
        display: inline-block;
        padding: 2pt 6pt;
        border-radius: 3pt;
        font-size: 8pt;
        font-weight: bold;
    }
    """

    # HTML complet avec en-tête
    html_full = f"""
    <!DOCTYPE html>
    <html lang="fr">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Cahier des Charges Final - Projet Honeypot</title>
    </head>
    <body>
        {html_content}
    </body>
    </html>
    """

    # Configuration des polices
    font_config = FontConfiguration()

    # Convertir HTML en PDF
    print(f"Conversion de {md_file} en PDF...")
    HTML(string=html_full).write_pdf(
        pdf_file,
        stylesheets=[CSS(string=css_style, font_config=font_config)],
        font_config=font_config
    )

    print(f"✅ PDF créé avec succès : {pdf_file}")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        md_file = sys.argv[1]
        pdf_file = sys.argv[2] if len(sys.argv) > 2 else md_file.replace('.md', '.pdf')
    else:
        # Par défaut, convertir le cahier des charges
        md_file = "/home/kali/Honeypot/docs/CAHIER_DES_CHARGES_FINAL.md"
        pdf_file = "/home/kali/Honeypot/docs/CAHIER_DES_CHARGES_FINAL.pdf"

    convert_markdown_to_pdf(md_file, pdf_file)
